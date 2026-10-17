// Package ratelimit implements a token bucket rate limiter.
//
// A bucket holds up to Burst tokens and is refilled at Rate tokens per
// second. Each event takes one token. When the bucket is empty, events
// are rejected (Allow) or delayed (Wait) until a token is available.
//
// This is the same algorithm as golang.org/x/time/rate, written out
// with only the standard library so every step is visible.
package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// ErrExceedsBurst is returned by WaitN when n can never be satisfied.
var ErrExceedsBurst = errors.New("ratelimit: n exceeds burst")

// Bucket is a token bucket. It is safe for concurrent use.
type Bucket struct {
	mu     sync.Mutex
	rate   float64 // tokens added per second
	burst  float64 // capacity
	tokens float64 // tokens currently available
	last   time.Time
	now    func() time.Time
}

// New returns a full bucket that refills at rate tokens per second and
// holds at most burst tokens. It panics if rate <= 0 or burst < 1,
// because those are programming errors (lesson 5.5).
func New(rate float64, burst int) *Bucket {
	if rate <= 0 || burst < 1 {
		panic(fmt.Sprintf("ratelimit.New: invalid rate %v or burst %d", rate, burst))
	}
	b := &Bucket{rate: rate, burst: float64(burst), tokens: float64(burst), now: time.Now}
	b.last = b.now()
	return b
}

// refill adds the tokens earned since the last call. b.mu must be held.
// Refilling lazily, on demand, means no background goroutine or timer
// is needed: the bucket costs nothing while idle.
func (b *Bucket) refill(now time.Time) {
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(b.burst, b.tokens+elapsed*b.rate)
		b.last = now
	}
}

// Allow reports whether one event may happen now, and takes a token if so.
func (b *Bucket) Allow() bool { return b.AllowN(1) }

// AllowN reports whether n events may happen now, taking n tokens if so.
func (b *Bucket) AllowN(n int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refill(b.now())
	if b.tokens >= float64(n) {
		b.tokens -= float64(n)
		return true
	}
	return false
}

// Tokens returns the number of tokens currently available.
func (b *Bucket) Tokens() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refill(b.now())
	return b.tokens
}

// Wait blocks until one token is available or ctx is done.
func (b *Bucket) Wait(ctx context.Context) error { return b.WaitN(ctx, 1) }

// WaitN blocks until n tokens are available or ctx is done.
//
// It reserves the tokens up front (the balance may go negative) and
// then sleeps for the time needed to earn them back. Reserving first
// keeps waiters fair: each one is scheduled after the previous ones.
func (b *Bucket) WaitN(ctx context.Context, n int) error {
	if float64(n) > b.burst {
		return fmt.Errorf("%w: n=%d burst=%v", ErrExceedsBurst, n, b.burst)
	}

	b.mu.Lock()
	now := b.now()
	b.refill(now)
	b.tokens -= float64(n)
	var delay time.Duration
	if b.tokens < 0 {
		delay = time.Duration(-b.tokens / b.rate * float64(time.Second))
	}
	b.mu.Unlock()

	if delay == 0 {
		return nil
	}
	// Do not even start waiting if the deadline is too close.
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(now.Add(delay)) {
		b.cancel(n)
		return fmt.Errorf("ratelimit: wait of %v would exceed context deadline: %w", delay, context.DeadlineExceeded)
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		b.cancel(n)
		return ctx.Err()
	}
}

// cancel returns n reserved tokens after a waiter gave up.
func (b *Bucket) cancel(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens = math.Min(b.burst, b.tokens+float64(n))
}
