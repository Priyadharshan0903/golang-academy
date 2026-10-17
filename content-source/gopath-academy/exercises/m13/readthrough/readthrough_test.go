package readthrough

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (f *fakeClock) Now() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.t }
func (f *fakeClock) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.t = f.t.Add(d)
}

func newClock() *fakeClock { return &fakeClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)} }

func TestHitMissExpiry(t *testing.T) {
	clock := newClock()
	var loads atomic.Int64
	c := New(func(ctx context.Context, key string) (string, error) {
		loads.Add(1)
		return "v:" + key, nil
	}, time.Minute, clock.Now)
	ctx := context.Background()

	for range 3 {
		if v, err := c.Get(ctx, "a"); v != "v:a" || err != nil {
			t.Fatalf("Get = %q, %v", v, err)
		}
	}
	if loads.Load() != 1 {
		t.Fatalf("3 Gets caused %d loads, want 1 (rule 1 and 2)", loads.Load())
	}
	clock.Advance(59 * time.Second)
	c.Get(ctx, "a")
	if loads.Load() != 1 {
		t.Fatalf("loaded again before the TTL ran out")
	}
	clock.Advance(2 * time.Second)
	c.Get(ctx, "a")
	if loads.Load() != 2 {
		t.Fatalf("after the TTL: %d loads, want 2 (rule 2)", loads.Load())
	}
}

func TestErrorsNotCached(t *testing.T) {
	var loads atomic.Int64
	c := New(func(ctx context.Context, key string) (string, error) {
		if loads.Add(1) == 1 {
			return "", errors.New("database unavailable")
		}
		return "ok", nil
	}, time.Minute, newClock().Now)

	if _, err := c.Get(context.Background(), "k"); err == nil {
		t.Fatal("first Get should return the loader's error")
	}
	if v, err := c.Get(context.Background(), "k"); v != "ok" || err != nil {
		t.Fatalf("second Get = %q, %v: errors must not be cached (rule 3)", v, err)
	}
}

func TestCoalescing(t *testing.T) {
	var loads atomic.Int64
	release := make(chan struct{})
	c := New(func(ctx context.Context, key string) (string, error) {
		loads.Add(1)
		<-release // hold the load open until every goroutine has called Get
		return "shared", nil
	}, time.Minute, newClock().Now)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := c.Get(context.Background(), "hot"); v != "shared" || err != nil {
				t.Errorf("Get = %q, %v", v, err)
			}
		}()
	}
	time.Sleep(100 * time.Millisecond) // let all 50 reach Get
	close(release)
	wg.Wait()
	if loads.Load() != 1 {
		t.Fatalf("50 concurrent misses caused %d loads, want 1 (rule 4)", loads.Load())
	}
}

func TestOtherKeysNotBlocked(t *testing.T) {
	block := make(chan struct{})
	c := New(func(ctx context.Context, key string) (string, error) {
		if key == "slow" {
			<-block
		}
		return key, nil
	}, time.Minute, newClock().Now)

	go c.Get(context.Background(), "slow")
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() { c.Get(context.Background(), "fast"); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal(`Get("fast") waited for the load of "slow" (rule 5)`)
	}
	close(block)
}
