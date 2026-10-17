// Package cache wraps a link.Store with a Redis cache (cache-aside,
// lesson 13.5). Links never change once created, which makes them
// ideal to cache: there is nothing to invalidate.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// Store is a link.Store that serves Get from Redis when it can.
type Store struct {
	next   link.Store
	rdb    *redis.Client
	ttl    time.Duration
	logger *slog.Logger
	group  singleflight.Group
}

// New wraps next. Entries expire after ttl.
func New(next link.Store, rdb *redis.Client, ttl time.Duration, logger *slog.Logger) *Store {
	return &Store{next: next, rdb: rdb, ttl: ttl, logger: logger}
}

func key(code string) string { return "shortener:link:" + code }

// Create writes to the underlying store only. The first Get fills the
// cache.
func (s *Store) Create(ctx context.Context, l link.Link) error {
	return s.next.Create(ctx, l)
}

// Delete removes the link from the underlying store, then from the
// cache. Deleting in this order means a concurrent Get can at worst
// re-cache the link just before it is deleted from the store; the
// cache entry is then removed right after.
func (s *Store) Delete(ctx context.Context, code string) error {
	if err := s.next.Delete(ctx, code); err != nil {
		return err
	}
	if err := s.rdb.Del(ctx, key(code)).Err(); err != nil {
		s.logger.Warn("cache delete failed", "code", code, "error", err)
	}
	return nil
}

// Get looks in Redis, then falls back to the underlying store. Redis
// errors are logged and otherwise ignored: the cache is an
// optimization, never a dependency.
func (s *Store) Get(ctx context.Context, code string) (link.Link, error) {
	b, err := s.rdb.Get(ctx, key(code)).Bytes()
	if err == nil {
		var l link.Link
		if err := json.Unmarshal(b, &l); err == nil {
			return l, nil
		}
		s.logger.Warn("corrupt cache entry", "code", code)
	} else if !errors.Is(err, redis.Nil) {
		s.logger.Warn("cache read failed", "code", code, "error", err)
	}

	// Concurrent misses for one code share a single store lookup.
	v, err, _ := s.group.Do(code, func() (any, error) {
		l, err := s.next.Get(context.WithoutCancel(ctx), code)
		if err != nil {
			return link.Link{}, err // misses are not cached
		}
		if b, err := json.Marshal(l); err == nil {
			if err := s.rdb.Set(ctx, key(code), b, s.ttl).Err(); err != nil {
				s.logger.Warn("cache write failed", "code", code, "error", err)
			}
		}
		return l, nil
	})
	return v.(link.Link), err
}
