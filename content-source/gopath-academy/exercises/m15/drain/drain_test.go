package drain

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

type harness struct {
	url    string
	ready  *Ready
	cancel context.CancelFunc
	done   chan error
}

func start(t *testing.T, h http.Handler, opts Options) *harness {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	hs := &harness{url: "http://" + ln.Addr().String(), ready: new(Ready), cancel: cancel, done: make(chan error, 1)}
	srv := &http.Server{Handler: h}
	go func() { hs.done <- Serve(ctx, srv, ln, hs.ready, opts) }()
	deadline := time.Now().Add(2 * time.Second)
	for !hs.ready.Load() {
		if time.Now().After(deadline) {
			t.Fatal("never became ready")
		}
		time.Sleep(5 * time.Millisecond)
	}
	return hs
}

func get(url string) (string, error) {
	c := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := c.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b), nil
}

func ok(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok") }

func TestUnreadyAtOnceButStillServing(t *testing.T) {
	hs := start(t, http.HandlerFunc(ok), Options{Delay: 300 * time.Millisecond, Timeout: time.Second})
	hs.cancel()
	time.Sleep(20 * time.Millisecond)
	if hs.ready.Load() {
		t.Fatal("still ready after cancel: load balancers keep sending traffic")
	}
	if body, err := get(hs.url); err != nil || body != "ok" {
		t.Fatalf("during the delay, requests must still be served: %q %v", body, err)
	}
	select {
	case err := <-hs.done:
		t.Fatalf("Serve returned before the delay was over: %v", err)
	default:
	}
	if err := <-hs.done; err != nil {
		t.Fatalf("Serve = %v, want nil after a clean shutdown", err)
	}
}

func TestInFlightRequestsFinish(t *testing.T) {
	started := make(chan struct{})
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(300 * time.Millisecond)
		io.WriteString(w, "finished")
	})
	hs := start(t, slow, Options{Delay: 0, Timeout: 2 * time.Second})
	result := make(chan string, 1)
	go func() { body, _ := get(hs.url); result <- body }()
	<-started
	hs.cancel()
	if body := <-result; body != "finished" {
		t.Fatalf("in-flight request got %q, want it to finish", body)
	}
	if err := <-hs.done; err != nil {
		t.Fatalf("Serve = %v", err)
	}
}

func TestShutdownTimeout(t *testing.T) {
	started := make(chan struct{})
	stuck := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(2 * time.Second)
	})
	hs := start(t, stuck, Options{Delay: 0, Timeout: 100 * time.Millisecond})
	go get(hs.url)
	<-started
	begin := time.Now()
	hs.cancel()
	err := <-hs.done
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Serve = %v, want context.DeadlineExceeded", err)
	}
	if d := time.Since(begin); d > time.Second {
		t.Fatalf("Serve took %v; it must give up after Timeout", d)
	}
}
