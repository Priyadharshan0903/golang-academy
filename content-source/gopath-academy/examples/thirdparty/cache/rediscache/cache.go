// Package rediscache holds the caching examples for lesson 13.5: a
// cache-aside layer over Redis, with request coalescing.
package rediscache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// [start cache]

// A Loader fetches a value from the source of truth, such as a database.
type Loader func(ctx context.Context, key string) (string, error)

// Cache implements cache-aside: look in Redis first, and on a miss,
// load from the source and store the result with a TTL.
type Cache struct {
	rdb    *redis.Client
	load   Loader
	ttl    time.Duration
	prefix string
	logger *slog.Logger

	Coalesce bool // share one load among concurrent misses for a key
	group    singleflight.Group
}

func New(rdb *redis.Client, prefix string, ttl time.Duration, load Loader, logger *slog.Logger) *Cache {
	return &Cache{rdb: rdb, load: load, ttl: ttl, prefix: prefix, logger: logger, Coalesce: true}
}

// Get returns the value for key. If Redis is unavailable, it falls back
// to the loader: a cache must make things faster, never less available.
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	v, err := c.rdb.Get(ctx, c.prefix+key).Result()
	switch {
	case err == nil:
		return v, nil // hit
	case errors.Is(err, redis.Nil):
		// miss: fall through
	default:
		c.logger.Warn("cache read failed", "key", key, "error", err)
	}

	if !c.Coalesce {
		return c.loadAndStore(ctx, key)
	}
	// Concurrent misses for the same key wait for one load (the
	// "thundering herd" or "cache stampede" problem).
	res, err, _ := c.group.Do(key, func() (any, error) {
		return c.loadAndStore(context.WithoutCancel(ctx), key)
	})
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func (c *Cache) loadAndStore(ctx context.Context, key string) (string, error) {
	v, err := c.load(ctx, key)
	if err != nil {
		return "", err
	}
	if err := c.rdb.Set(ctx, c.prefix+key, v, c.ttl).Err(); err != nil {
		c.logger.Warn("cache write failed", "key", key, "error", err)
	}
	return v, nil
}

// Invalidate removes key, for example after the source changed.
func (c *Cache) Invalidate(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, c.prefix+key).Err()
}

// [end cache]
