//go:build !solution

// Package readthrough practices the caching patterns of lesson 13.5
// without needing Redis: an in-memory read-through cache with expiry
// and request coalescing.
package readthrough

import (
	"context"
	"time"
)

// Loader fetches the value for key from the slow source of truth.
type Loader func(ctx context.Context, key string) (string, error)

// Cache sits in front of a Loader. It must:
//
//  1. return a cached value without calling the loader (a hit);
//  2. call the loader on a miss, and cache the value for ttl;
//  3. never cache errors: the next Get tries again;
//  4. call the loader only once when many goroutines miss the same key
//     at the same time; all of them get that one result;
//  5. not block Gets for other keys while a load is in progress.
//
// now is the clock; tests pass a fake one to control expiry.
type Cache struct {
	load Loader
	ttl  time.Duration
	now  func() time.Time
	// TODO: add the fields you need.
}

// New returns an empty cache.
func New(load Loader, ttl time.Duration, now func() time.Time) *Cache {
	return &Cache{load: load, ttl: ttl, now: now}
}

// Get returns the value for key.
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	// TODO: implement the five rules above. This placeholder never caches.
	return c.load(ctx, key)
}
