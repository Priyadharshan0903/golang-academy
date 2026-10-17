//go:build solution

// Package readthrough practices the caching patterns of lesson 13.5
// without needing Redis: an in-memory read-through cache with expiry
// and request coalescing.
package readthrough

import (
	"context"
	"sync"
	"time"
)

// Loader fetches the value for key from the slow source of truth.
type Loader func(ctx context.Context, key string) (string, error)

type entry struct {
	value   string
	expires time.Time
}

// call is a load in progress. Goroutines that miss the same key wait on
// done and then read the shared result.
type call struct {
	done  chan struct{}
	value string
	err   error
}

// Cache sits in front of a Loader. See the starter for the rules.
type Cache struct {
	load Loader
	ttl  time.Duration
	now  func() time.Time

	mu       sync.Mutex // guards entries and inflight, never held during a load
	entries  map[string]entry
	inflight map[string]*call
}

// New returns an empty cache.
func New(load Loader, ttl time.Duration, now func() time.Time) *Cache {
	return &Cache{load: load, ttl: ttl, now: now, entries: map[string]entry{}, inflight: map[string]*call{}}
}

// Get returns the value for key.
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	c.mu.Lock()
	if e, ok := c.entries[key]; ok && c.now().Before(e.expires) {
		c.mu.Unlock()
		return e.value, nil // rule 1
	}
	if cl, ok := c.inflight[key]; ok {
		c.mu.Unlock()
		select { // rule 4: wait for the load already in progress
		case <-cl.done:
			return cl.value, cl.err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	cl := &call{done: make(chan struct{})}
	c.inflight[key] = cl
	c.mu.Unlock() // rule 5: do not hold the lock while loading

	cl.value, cl.err = c.load(ctx, key)

	c.mu.Lock()
	if cl.err == nil { // rule 3: only successes are cached
		c.entries[key] = entry{value: cl.value, expires: c.now().Add(c.ttl)} // rule 2
	}
	delete(c.inflight, key)
	c.mu.Unlock()
	close(cl.done)
	return cl.value, cl.err
}
