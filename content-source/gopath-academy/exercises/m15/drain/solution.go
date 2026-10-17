//go:build solution

// Package drain practices the shutdown sequence of a service running in
// Kubernetes (lesson 15.2).
package drain

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// Options configure Serve. See the starter for details.
type Options struct {
	Delay   time.Duration
	Timeout time.Duration
}

// Ready reports whether the service should receive traffic.
type Ready = atomic.Bool

// Serve serves srv on ln until ctx is cancelled, then shuts down
// gracefully. See the starter for the specification.
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, ready *Ready, opts Options) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()
	ready.Store(true)

	select {
	case err := <-serveErr: // the server failed on its own
		ready.Store(false)
		return err
	case <-ctx.Done():
	}

	ready.Store(false) // step 2: stop attracting new traffic
	t := time.NewTimer(opts.Delay)
	select { // step 3: keep serving, unless the server dies meanwhile
	case <-t.C:
	case err := <-serveErr:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()
	err := srv.Shutdown(shutdownCtx) // step 4
	if serr := <-serveErr; !errors.Is(serr, http.ErrServerClosed) {
		return errors.Join(err, serr)
	}
	return err // step 5: nil, or context.DeadlineExceeded if requests ran too long
}
