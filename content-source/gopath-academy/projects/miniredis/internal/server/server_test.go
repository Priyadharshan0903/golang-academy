package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/resp"
	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/store"
)

// startServer runs a server on a random local port and stops it at the
// end of the test.
func startServer(t *testing.T, aof Appender) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	srv := &Server{Store: store.New(), AOF: aof}
	done := make(chan error, 1)
	go func() { done <- srv.Serve(ctx, ln) }()
	t.Cleanup(func() {
		cancel()
		if err := <-done; err != nil {
			t.Errorf("Serve: %v", err)
		}
	})
	return ln.Addr().String()
}

// client is a minimal Redis client for tests.
type client struct {
	conn net.Conn
	r    *bufio.Reader
	w    *resp.Writer
}

func dial(t *testing.T, addr string) *client {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return &client{conn: c, r: bufio.NewReader(c), w: resp.NewWriter(c)}
}

// do sends a command and returns the raw reply line(s) as one string.
func (c *client) do(t *testing.T, args ...string) string {
	t.Helper()
	c.conn.SetDeadline(time.Now().Add(2 * time.Second))
	c.w.Array(args)
	if err := c.w.Flush(); err != nil {
		t.Fatal(err)
	}
	line, err := c.r.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(line, "$") && line != "$-1\r\n" {
		body, _ := c.r.ReadString('\n')
		return strings.TrimSpace(body)
	}
	return strings.TrimSpace(line)
}

func TestCommands(t *testing.T) {
	c := dial(t, startServer(t, nil))
	steps := []struct {
		args []string
		want string
	}{
		{[]string{"PING"}, "+PONG"},
		{[]string{"ping", "hello"}, "hello"},
		{[]string{"SET", "name", "Ana Maria"}, "+OK"},
		{[]string{"GET", "name"}, "Ana Maria"},
		{[]string{"GET", "missing"}, "$-1"},
		{[]string{"INCR", "hits"}, ":1"},
		{[]string{"INCR", "hits"}, ":2"},
		{[]string{"INCR", "name"}, "-ERR value is not an integer or out of range"},
		{[]string{"EXISTS", "name", "hits", "nope"}, ":2"},
		{[]string{"SET", "tmp", "x", "EX", "100"}, "+OK"},
		{[]string{"TTL", "tmp"}, ":100"},
		{[]string{"TTL", "name"}, ":-1"},
		{[]string{"DEL", "name", "nope"}, ":1"},
		{[]string{"DBSIZE"}, ":2"},
		{[]string{"GET"}, "-ERR wrong number of arguments for 'get' command"},
		{[]string{"FLY"}, "-ERR unknown command 'FLY'"},
	}
	for _, s := range steps {
		if got := c.do(t, s.args...); got != s.want {
			t.Errorf("%q -> %q, want %q", s.args, got, s.want)
		}
	}
}

func TestInlineCommand(t *testing.T) {
	addr := startServer(t, nil)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	fmt.Fprint(conn, "PING\r\n") // what you would type in telnet
	line, _ := bufio.NewReader(conn).ReadString('\n')
	if line != "+PONG\r\n" {
		t.Errorf("inline PING -> %q", line)
	}
}

func TestManyConcurrentClients(t *testing.T) {
	addr := startServer(t, nil)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := dial(t, addr)
			for range 50 {
				c.do(t, "INCR", "shared")
			}
		}()
	}
	wg.Wait()
	if got := dial(t, addr).do(t, "GET", "shared"); got != "1000" {
		t.Errorf("shared counter = %s, want 1000", got)
	}
}

type recorder struct {
	mu   sync.Mutex
	cmds []string
}

func (r *recorder) Append(args []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cmds = append(r.cmds, strings.Join(args, " "))
	return nil
}

func TestOnlyWritesArePersisted(t *testing.T) {
	rec := &recorder{}
	c := dial(t, startServer(t, rec))
	c.do(t, "SET", "a", "1")
	c.do(t, "GET", "a")
	c.do(t, "DEL", "nope") // deleted nothing: not persisted
	c.do(t, "INCR", "n")
	rec.mu.Lock()
	defer rec.mu.Unlock()
	want := []string{"SET a 1", "INCR n"}
	if strings.Join(rec.cmds, "|") != strings.Join(want, "|") {
		t.Errorf("persisted %q, want %q", rec.cmds, want)
	}
}

func TestShutdownClosesClients(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- (&Server{Store: store.New()}).Serve(ctx, ln) }()

	c := dial(t, ln.Addr().String())
	c.do(t, "PING") // connected and idle

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return: idle clients blocked shutdown")
	}
}
