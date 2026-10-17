package ratelimit

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeClock is a manually advanced clock for deterministic tests.
type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func newTestBucket(rate float64, burst int) (*Bucket, *fakeClock) {
	clock := &fakeClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	b := New(rate, burst)
	b.now = clock.Now
	b.last = clock.Now()
	return b, clock
}

func TestBurstThenEmpty(t *testing.T) {
	b, _ := newTestBucket(10, 3)
	for i := range 3 {
		if !b.Allow() {
			t.Fatalf("event %d should be allowed by the burst", i)
		}
	}
	if b.Allow() {
		t.Error("4th event should be rejected: bucket is empty")
	}
}

func TestRefill(t *testing.T) {
	b, clock := newTestBucket(10, 5) // one token every 100ms
	for b.Allow() {
	}
	clock.Advance(250 * time.Millisecond)
	if got := b.Tokens(); got < 2.49 || got > 2.51 {
		t.Errorf("after 250ms: %.2f tokens, want 2.5", got)
	}
	clock.Advance(time.Hour)
	if got := b.Tokens(); got != 5 {
		t.Errorf("tokens never exceed burst: got %v, want 5", got)
	}
}

func TestAllowN(t *testing.T) {
	b, _ := newTestBucket(1, 10)
	if !b.AllowN(7) || b.AllowN(4) || !b.AllowN(3) {
		t.Error("AllowN accounting is wrong")
	}
}

func TestWaitImmediate(t *testing.T) {
	b, _ := newTestBucket(1, 1)
	start := time.Now()
	if err := b.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Error("Wait should return immediately when a token is available")
	}
}

func TestWaitDelays(t *testing.T) {
	b := New(50, 1) // real clock: one token per 20ms
	b.Allow()
	start := time.Now()
	if err := b.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d < 15*time.Millisecond {
		t.Errorf("Wait returned after %v, want about 20ms", d)
	}
}

func TestWaitRespectsDeadline(t *testing.T) {
	b := New(1, 1) // a new token takes a full second
	b.Allow()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := b.Wait(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Error("Wait should fail fast when the deadline is too close")
	}
	if got := b.Tokens(); got < -0.01 {
		t.Errorf("the failed waiter's reservation was not returned: %.2f tokens", got)
	}
}

func TestWaitNExceedsBurst(t *testing.T) {
	b := New(10, 2)
	if err := b.WaitN(context.Background(), 3); !errors.Is(err, ErrExceedsBurst) {
		t.Errorf("err = %v, want ErrExceedsBurst", err)
	}
}

func TestConcurrentAllow(t *testing.T) {
	b, _ := newTestBucket(1, 100) // frozen clock: exactly 100 tokens
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range 500 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b.Allow() {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 100 {
		t.Errorf("allowed %d events, want exactly 100", allowed.Load())
	}
}

func TestNewPanicsOnBadConfig(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("New(0, 1) should panic")
		}
	}()
	New(0, 1)
}

func TestKeyed(t *testing.T) {
	clock := &fakeClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	k := NewKeyed(1, 2)
	k.now = clock.Now

	if !k.Allow("a") || !k.Allow("a") || k.Allow("a") {
		t.Error("key a should get exactly its burst of 2")
	}
	if !k.Allow("b") {
		t.Error("key b has its own bucket")
	}
	clock.Advance(time.Minute)
	k.Allow("b")
	if removed := k.Prune(30 * time.Second); removed != 1 || k.Len() != 1 {
		t.Errorf("Prune removed %d, Len %d; want 1 and 1 (only a was idle)", removed, k.Len())
	}
}

func BenchmarkAllow(b *testing.B) {
	lim := New(1e9, 1<<30)
	for b.Loop() {
		lim.Allow()
	}
}

func BenchmarkAllowParallel(b *testing.B) {
	lim := New(1e9, 1<<30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lim.Allow()
		}
	})
}
