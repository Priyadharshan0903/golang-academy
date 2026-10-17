// Package store is the in-memory keyspace: string values with optional
// expiry, safe for concurrent use by many client connections.
package store

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"
)

// ErrNotInteger is returned by Incr when the value is not an integer.
var ErrNotInteger = errors.New("value is not an integer or out of range")

type entry struct {
	value   string
	expires time.Time // zero means "never"
}

func (e entry) expired(now time.Time) bool {
	return !e.expires.IsZero() && !now.Before(e.expires)
}

// Store is a concurrency-safe map of string keys to string values.
type Store struct {
	mu   sync.RWMutex
	data map[string]entry
	now  func() time.Time
}

// New returns an empty store.
func New() *Store { return &Store{data: make(map[string]entry), now: time.Now} }

// Get returns the value of key. Expired keys are treated as missing
// (and removed lazily by later writes or by the janitor).
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.data[key]
	if !ok || e.expired(s.now()) {
		return "", false
	}
	return e.value, true
}

// Set stores value under key. A ttl of 0 means no expiry.
func (s *Store) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := entry{value: value}
	if ttl > 0 {
		e.expires = s.now().Add(ttl)
	}
	s.data[key] = e
}

// Del removes keys and returns how many existed.
func (s *Store) Del(keys ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	n := 0
	for _, k := range keys {
		if e, ok := s.data[k]; ok {
			if !e.expired(now) {
				n++
			}
			delete(s.data, k)
		}
	}
	return n
}

// Exists returns how many of keys exist (duplicates count each time, like Redis).
func (s *Store) Exists(keys ...string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	n := 0
	for _, k := range keys {
		if e, ok := s.data[k]; ok && !e.expired(now) {
			n++
		}
	}
	return n
}

// Incr adds 1 to the integer value of key (missing keys start at 0) and
// returns the new value. The key keeps its expiry.
func (s *Store) Incr(key string) (int64, error) {
	s.mu.Lock() // the read-modify-write must be one critical section
	defer s.mu.Unlock()
	e, ok := s.data[key]
	if !ok || e.expired(s.now()) {
		e = entry{value: "0"}
	}
	n, err := strconv.ParseInt(e.value, 10, 64)
	if err != nil || n == 1<<63-1 {
		return 0, ErrNotInteger
	}
	n++
	e.value = strconv.FormatInt(n, 10)
	s.data[key] = e
	return n, nil
}

// Expire sets a ttl on an existing key and reports whether the key existed.
func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	e, ok := s.data[key]
	if !ok || e.expired(now) {
		return false
	}
	e.expires = now.Add(ttl)
	s.data[key] = e
	return true
}

// TTL returns the remaining time to live. It returns -2 for a missing
// key and -1 for a key without expiry, matching Redis' TTL command.
func (s *Store) TTL(key string) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	e, ok := s.data[key]
	switch {
	case !ok || e.expired(now):
		return -2
	case e.expires.IsZero():
		return -1
	default:
		return e.expires.Sub(now)
	}
}

// Len returns the number of keys, including expired ones not yet removed.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// RemoveExpired deletes every expired key and returns how many it removed.
func (s *Store) RemoveExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	n := 0
	for k, e := range s.data {
		if e.expired(now) {
			delete(s.data, k)
			n++
		}
	}
	return n
}

// RunJanitor removes expired keys every interval until ctx is done.
// Without it, keys that expire and are never read again would stay in
// memory forever.
func (s *Store) RunJanitor(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.RemoveExpired()
		}
	}
}
