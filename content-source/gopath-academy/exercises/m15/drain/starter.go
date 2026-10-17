//go:build !solution

// Package drain practices the shutdown sequence of a service running in
// Kubernetes (lesson 15.2).
package drain

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// Options configure Serve.
type Options struct {
	// Delay is how long to keep serving after ctx is cancelled, while
	// reporting not ready, so load balancers stop sending new traffic.
	Delay time.Duration
	// Timeout bounds how long Shutdown waits for in-flight requests.
	Timeout time.Duration
}

// Ready reports whether the service should receive traffic. Serve sets
// it to true when it starts serving and to false as soon as ctx is
// cancelled. A /readyz handler would read it.
type Ready = atomic.Bool

// Serve serves srv on ln until ctx is cancelled, then shuts down
// gracefully:
//
//  1. set ready to true once serving;
//  2. when ctx is cancelled, set ready to false immediately;
//  3. keep serving normally for opts.Delay;
//  4. call srv.Shutdown with a context that times out after
//     opts.Timeout, so in-flight requests can finish;
//  5. return nil after a clean shutdown, or the error from Serve or
//     Shutdown (never http.ErrServerClosed).
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, ready *Ready, opts Options) error {
	// TODO: implement the five steps above.
	return srv.Serve(ln)
}
