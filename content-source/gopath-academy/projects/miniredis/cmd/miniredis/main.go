// Command miniredis is a small Redis-compatible server, built as the
// Module 8 project. It supports PING, ECHO, GET, SET [EX], DEL, EXISTS,
// INCR, EXPIRE, TTL and DBSIZE, persists writes to an append-only file,
// and works with redis-cli.
//
//	go run ./projects/miniredis/cmd/miniredis -addr :6380 -aof data.aof
//	redis-cli -p 6380 set greeting hello
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/aof"
	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/server"
	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "miniredis:", err)
		os.Exit(1)
	}
}

func run() error {
	addr := flag.String("addr", "127.0.0.1:6380", "listen address")
	aofPath := flag.String("aof", "", "append-only file (empty disables persistence)")
	fsyncAlways := flag.Bool("fsync-always", false, "fsync after every write instead of every second")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st := store.New()
	srv := &server.Server{Store: st, Logger: logger}

	if *aofPath != "" {
		// Replay into the store directly, so replayed commands are not re-appended.
		replayer := &server.Server{Store: st}
		n, err := aof.Replay(*aofPath, func(args []string) error {
			return replayer.Apply(args)
		})
		if err != nil {
			return err
		}
		logger.Info("aof replayed", "commands", n)

		log, err := aof.Open(*aofPath, *fsyncAlways)
		if err != nil {
			return err
		}
		defer log.Close()
		srv.AOF = log
		stopSync := make(chan struct{})
		defer close(stopSync)
		go log.SyncEvery(time.Second, stopSync)
	}

	go st.RunJanitor(ctx, 100*time.Millisecond)

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return err
	}
	logger.Info("listening", "addr", ln.Addr().String())
	err = srv.Serve(ctx, ln)
	logger.Info("stopped")
	return err
}
