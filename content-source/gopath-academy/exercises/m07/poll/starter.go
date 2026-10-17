//go:build !solution

// Package poll practices context: waiting with deadlines and retrying
// without ignoring cancellation.
package poll

import (
	"context"
	"time"
)

// WaitFor calls cond immediately and then once per interval until it
// returns true (WaitFor returns nil) or ctx is done (WaitFor returns an
// error wrapping ctx.Err(), so errors.Is(err, context.DeadlineExceeded)
// works).
//
// Use a time.Ticker and stop it before returning.
func WaitFor(ctx context.Context, interval time.Duration, cond func() bool) error {
	// TODO: implement me.
	return nil
}

// Retry calls f up to attempts times until it returns nil. Between
// attempts it waits, starting at backoff and doubling each time. If ctx
// is done while waiting, Retry stops immediately and returns an error
// wrapping ctx.Err(). If every attempt fails, it returns the last error
// from f (wrapped is fine).
//
// f receives ctx so it can stop its own work early.
func Retry(ctx context.Context, attempts int, backoff time.Duration, f func(context.Context) error) error {
	// TODO: implement me.
	return nil
}
