package cache

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link/linktest"
	"github.com/redis/go-redis/v9"
)

func redisClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR not set; see projects/shortener/README.md")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { rdb.Close() })
	rdb.FlushDB(context.Background()) // tests use their own Redis
	return rdb
}

func TestCachedStoreContract(t *testing.T) {
	rdb := redisClient(t)
	linktest.Run(t, func(t *testing.T) link.Store {
		rdb.FlushDB(context.Background())
		return New(link.NewMemoryStore(), rdb, time.Minute, slog.New(slog.DiscardHandler))
	})
}

// countingStore counts Get calls that reach the underlying store.
type countingStore struct {
	link.Store
	gets atomic.Int64
}

func (c *countingStore) Get(ctx context.Context, code string) (link.Link, error) {
	c.gets.Add(1)
	time.Sleep(20 * time.Millisecond)
	return c.Store.Get(ctx, code)
}

func TestCacheHitsAndCoalescing(t *testing.T) {
	under := &countingStore{Store: link.NewMemoryStore()}
	s := New(under, redisClient(t), time.Minute, slog.New(slog.DiscardHandler))
	ctx := t.Context()
	s.Create(ctx, link.Link{Code: "hot", URL: "https://go.dev", CreatedAt: time.Now().UTC()})

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			if l, err := s.Get(ctx, "hot"); err != nil || l.URL != "https://go.dev" {
				t.Error(l, err)
			}
		})
	}
	wg.Wait()
	for range 10 {
		s.Get(ctx, "hot")
	}
	if n := under.gets.Load(); n != 1 {
		t.Fatalf("60 Gets reached the store %d times, want 1", n)
	}
}

func TestRedisDownFallsBack(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 50 * time.Millisecond, MaxRetries: -1})
	defer rdb.Close()
	mem := link.NewMemoryStore()
	mem.Create(t.Context(), link.Link{Code: "abc", URL: "https://go.dev", CreatedAt: time.Now().UTC()})
	s := New(mem, rdb, time.Minute, slog.New(slog.DiscardHandler))
	if l, err := s.Get(t.Context(), "abc"); err != nil || l.URL != "https://go.dev" {
		t.Fatalf("Get = %+v, %v", l, err)
	}
}
