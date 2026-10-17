package shutdown

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// [start server]

// serve runs an HTTP server until ctx is cancelled (in a real program,
// by SIGINT or SIGTERM via signal.NotifyContext), then shuts down
// gracefully: stop accepting, let in-flight requests finish, give up
// after a deadline.
func serve(ctx context.Context, ln net.Listener, handler http.Handler) error {
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh: // the server failed on its own
		return err
	case <-ctx.Done(): // asked to stop
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// [end server]

// [start demo]

func Example_gracefulShutdown() {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	ctx, stop := context.WithCancel(context.Background())

	started := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(50 * time.Millisecond) // a slow, in-flight request
		fmt.Fprint(w, "finished")
	})

	done := make(chan error, 1)
	go func() { done <- serve(ctx, ln, handler) }()

	var wg sync.WaitGroup
	wg.Add(1)
	var body string
	go func() {
		defer wg.Done()
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			body = err.Error()
			return
		}
		defer resp.Body.Close()
		b := make([]byte, 64)
		n, _ := resp.Body.Read(b)
		body = string(b[:n])
	}()

	<-started
	stop() // shutdown begins while the request is still running

	wg.Wait()
	fmt.Println("in-flight request:", body)
	fmt.Println("serve returned:", <-done)

	// Output:
	// in-flight request: finished
	// serve returned: <nil>
}

// [end demo]
