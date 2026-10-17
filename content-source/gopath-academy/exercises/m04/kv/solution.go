//go:build solution

// Package kv practices interfaces: one small interface, several
// implementations, and decorators that wrap any implementation.
package kv

// Store is a minimal key-value store.
type Store interface {
	Get(key string) (value string, ok bool)
	Set(key, value string)
}

// MemStore is an in-memory Store.
type MemStore struct {
	data map[string]string
}

// NewMemStore returns an empty, ready to use MemStore.
func NewMemStore() *MemStore {
	return &MemStore{data: make(map[string]string)}
}

// Get returns the value for key and whether it was present.
func (m *MemStore) Get(key string) (string, bool) {
	v, ok := m.data[key]
	return v, ok
}

// Set stores value under key.
func (m *MemStore) Set(key, value string) { m.data[key] = value }

// Prefixed wraps another Store and adds Prefix to every key.
type Prefixed struct {
	Prefix string
	Inner  Store
}

// Get looks up Prefix+key in the inner store.
func (p Prefixed) Get(key string) (string, bool) { return p.Inner.Get(p.Prefix + key) }

// Set stores value under Prefix+key in the inner store.
func (p Prefixed) Set(key, value string) { p.Inner.Set(p.Prefix+key, value) }

// Counting wraps another Store and counts calls to Get and Set.
type Counting struct {
	Inner              Store
	Hits, Misses, Sets int
}

// Get forwards to the inner store and records a hit or a miss.
func (c *Counting) Get(key string) (string, bool) {
	v, ok := c.Inner.Get(key)
	if ok {
		c.Hits++
	} else {
		c.Misses++
	}
	return v, ok
}

// Set forwards to the inner store and records the call.
func (c *Counting) Set(key, value string) {
	c.Sets++
	c.Inner.Set(key, value)
}
