//go:build solution

// Package poll practices context: waiting with deadlines and retrying
// without ignoring cancellation.
package poll

import (
	"context"
	"fmt"
	"time"
)

// WaitFor polls cond until it is true or ctx is done.
func WaitFor(ctx context.Context, interval time.Duration, cond func() bool) error {
	if cond() {
		return nil
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for condition: %w", ctx.Err())
		case <-ticker.C:
			if cond() {
				return nil
			}
		}
	}
}

// Retry calls f with exponential backoff, respecting ctx.
func Retry(ctx context.Context, attempts int, backoff time.Duration, f func(context.Context) error) error {
	var err error
	for i := range attempts {
		if err = f(ctx); err == nil {
			return nil
		}
		if i == attempts-1 {
			break
		}
		timer := time.NewTimer(backoff << i)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("retry cancelled after %d attempts: %w", i+1, ctx.Err())
		case <-timer.C:
		}
	}
	return fmt.Errorf("all %d attempts failed: %w", attempts, err)
}
