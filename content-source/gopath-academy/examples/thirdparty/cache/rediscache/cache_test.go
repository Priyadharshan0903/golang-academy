package rediscache

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func client(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set; see examples/thirdparty/README.md")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

// slowSource simulates a database: 50 ms per load, and counts loads.
type slowSource struct{ loads atomic.Int64 }

func (s *slowSource) Load(ctx context.Context, key string) (string, error) {
	s.loads.Add(1)
	time.Sleep(50 * time.Millisecond)
	return "value of " + key, nil
}

func prefix(t *testing.T) string { return fmt.Sprintf("test:%s:%d:", t.Name(), time.Now().UnixNano()) }

func TestCacheAside(t *testing.T) {
	src := &slowSource{}
	c := New(client(t), prefix(t), 200*time.Millisecond, src.Load, slog.New(slog.DiscardHandler))
	ctx := t.Context()

	for i, want := range []int64{1, 1, 1} { // first call loads, the next two hit
		start := time.Now()
		v, err := c.Get(ctx, "user:42")
		if err != nil || v != "value of user:42" {
			t.Fatal(v, err)
		}
		t.Logf("call %d: %v, loads so far %d", i+1, time.Since(start).Round(100*time.Microsecond), src.loads.Load())
		if src.loads.Load() != want {
			t.Fatalf("loads = %d, want %d", src.loads.Load(), want)
		}
	}

	time.Sleep(300 * time.Millisecond) // past the TTL
	c.Get(ctx, "user:42")
	if src.loads.Load() != 2 {
		t.Fatalf("after TTL: loads = %d, want 2", src.loads.Load())
	}

	c.Invalidate(ctx, "user:42")
	c.Get(ctx, "user:42")
	if src.loads.Load() != 3 {
		t.Fatalf("after Invalidate: loads = %d, want 3", src.loads.Load())
	}
}

// [start stampede]

func TestStampede(t *testing.T) {
	rdb := client(t)
	for _, coalesce := range []bool{false, true} {
		src := &slowSource{}
		c := New(rdb, prefix(t), time.Minute, src.Load, slog.New(slog.DiscardHandler))
		c.Coalesce = coalesce

		var wg sync.WaitGroup
		for range 100 { // 100 requests for the same cold key at once
			wg.Go(func() {
				if _, err := c.Get(t.Context(), "popular"); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		t.Logf("coalesce=%-5v 100 concurrent misses caused %d loads", coalesce, src.loads.Load())
		if coalesce && src.loads.Load() != 1 {
			t.Fatalf("with coalescing, want exactly 1 load, got %d", src.loads.Load())
		}
	}
}

// [end stampede]

// [start down]

func TestRedisDown(t *testing.T) {
	// Nothing listens on this port: every Redis call fails quickly.
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 100 * time.Millisecond, MaxRetries: -1})
	defer rdb.Close()
	src := &slowSource{}
	c := New(rdb, "x:", time.Minute, src.Load, slog.New(slog.DiscardHandler))

	v, err := c.Get(t.Context(), "k")
	if err != nil || v != "value of k" {
		t.Fatalf("Get with Redis down = %q, %v; want the loaded value", v, err)
	}
}

// [end down]
