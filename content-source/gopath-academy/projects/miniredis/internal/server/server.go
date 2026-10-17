// Package server accepts TCP connections and executes Redis commands
// against a store, one goroutine per client.
package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/resp"
	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/store"
)

// Appender persists write commands. *aof.Log satisfies it.
type Appender interface {
	Append(args []string) error
}

// Server is a tiny Redis-compatible server.
type Server struct {
	Store  *store.Store
	AOF    Appender     // optional
	Logger *slog.Logger // optional

	mu    sync.Mutex
	conns map[net.Conn]struct{}
	wg    sync.WaitGroup
}

// Serve accepts connections on ln until ctx is cancelled, then closes
// every client connection and waits for their goroutines to finish.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	if s.Logger == nil {
		s.Logger = slog.New(slog.DiscardHandler)
	}
	s.mu.Lock()
	s.conns = make(map[net.Conn]struct{})
	s.mu.Unlock()

	// Closing the listener is what unblocks Accept on shutdown.
	stop := context.AfterFunc(ctx, func() { ln.Close() })
	defer stop()

	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break // shutting down
			}
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				time.Sleep(10 * time.Millisecond) // temporary: back off and retry
				continue
			}
			return err
		}
		s.track(conn, true)
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer s.track(conn, false)
			s.handle(conn)
		}()
	}

	s.mu.Lock()
	for c := range s.conns {
		c.Close() // unblocks each handler's read
	}
	s.mu.Unlock()
	s.wg.Wait()
	return nil
}

func (s *Server) track(c net.Conn, add bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if add {
		s.conns[c] = struct{}{}
	} else {
		delete(s.conns, c)
		c.Close()
	}
}

// handle serves one client until it disconnects.
func (s *Server) handle(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil { // a bug in one command must not crash the server
			s.Logger.Error("panic in connection handler", "remote", conn.RemoteAddr(), "panic", r)
		}
	}()
	r := resp.NewReader(conn)
	w := resp.NewWriter(conn)
	for {
		args, err := r.ReadCommand()
		if err != nil {
			if errors.Is(err, resp.ErrProtocol) {
				w.Error("ERR " + err.Error())
				w.Flush()
			}
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				s.Logger.Debug("connection closed", "remote", conn.RemoteAddr(), "err", err)
			}
			return
		}
		if len(args) == 0 {
			continue
		}
		s.execute(args, w)
		if err := w.Flush(); err != nil {
			return
		}
	}
}

// execute runs one command and writes its reply.
func (s *Server) execute(args []string, w *resp.Writer) {
	cmd := strings.ToUpper(args[0])
	argc := len(args) - 1
	wrongArgs := func() { w.Error(fmt.Sprintf("ERR wrong number of arguments for '%s' command", strings.ToLower(cmd))) }

	switch cmd {
	case "PING":
		switch argc {
		case 0:
			w.Simple("PONG")
		case 1:
			w.Bulk(args[1])
		default:
			wrongArgs()
		}
	case "ECHO":
		if argc != 1 {
			wrongArgs()
			return
		}
		w.Bulk(args[1])
	case "GET":
		if argc != 1 {
			wrongArgs()
			return
		}
		if v, ok := s.Store.Get(args[1]); ok {
			w.Bulk(v)
		} else {
			w.Null()
		}
	case "SET":
		if argc != 2 && argc != 4 {
			wrongArgs()
			return
		}
		var ttl time.Duration
		if argc == 4 {
			if !strings.EqualFold(args[3], "EX") {
				w.Error("ERR syntax error")
				return
			}
			secs, err := strconv.Atoi(args[4])
			if err != nil || secs <= 0 {
				w.Error("ERR invalid expire time in 'set' command")
				return
			}
			ttl = time.Duration(secs) * time.Second
		}
		s.Store.Set(args[1], args[2], ttl)
		s.persist(args)
		w.Simple("OK")
	case "DEL":
		if argc < 1 {
			wrongArgs()
			return
		}
		n := s.Store.Del(args[1:]...)
		if n > 0 {
			s.persist(args)
		}
		w.Int(int64(n))
	case "EXISTS":
		if argc < 1 {
			wrongArgs()
			return
		}
		w.Int(int64(s.Store.Exists(args[1:]...)))
	case "INCR":
		if argc != 1 {
			wrongArgs()
			return
		}
		n, err := s.Store.Incr(args[1])
		if err != nil {
			w.Error("ERR " + err.Error())
			return
		}
		s.persist(args)
		w.Int(n)
	case "EXPIRE":
		if argc != 2 {
			wrongArgs()
			return
		}
		secs, err := strconv.Atoi(args[2])
		if err != nil {
			w.Error("ERR value is not an integer or out of range")
			return
		}
		if s.Store.Expire(args[1], time.Duration(secs)*time.Second) {
			s.persist(args)
			w.Int(1)
		} else {
			w.Int(0)
		}
	case "TTL":
		if argc != 1 {
			wrongArgs()
			return
		}
		ttl := s.Store.TTL(args[1])
		if ttl < 0 {
			w.Int(int64(ttl)) // -1 or -2
		} else {
			w.Int(int64((ttl + time.Second - 1) / time.Second)) // round up like Redis
		}
	case "DBSIZE":
		w.Int(int64(s.Store.Len()))
	case "COMMAND": // redis-cli asks for this on connect; an empty reply is fine
		w.Array(nil)
	default:
		w.Error(fmt.Sprintf("ERR unknown command '%s'", args[0]))
	}
}

func (s *Server) persist(args []string) {
	if s.AOF == nil {
		return
	}
	if err := s.AOF.Append(args); err != nil {
		s.Logger.Error("aof append failed", "err", err)
	}
}
