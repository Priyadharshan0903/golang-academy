//go:build !solution

// Package kv practices interfaces: one small interface, several
// implementations, and decorators that wrap any implementation.
package kv

// Store is a minimal key-value store.
type Store interface {
	Get(key string) (value string, ok bool)
	Set(key, value string)
}

// MemStore is an in-memory Store. Its zero value is NOT ready to use;
// create one with NewMemStore.
type MemStore struct {
	// TODO: add the fields you need.
}

// NewMemStore returns an empty, ready to use MemStore.
func NewMemStore() *MemStore {
	// TODO: implement me.
	return &MemStore{}
}

// Get returns the value for key and whether it was present.
func (m *MemStore) Get(key string) (string, bool) {
	// TODO: implement me.
	return "", false
}

// Set stores value under key, replacing any previous value.
func (m *MemStore) Set(key, value string) {
	// TODO: implement me.
}

// Prefixed wraps another Store and adds Prefix to every key, so
// several users can share one Store without key collisions.
//
//	p := Prefixed{Prefix: "user:", Inner: mem}
//	p.Set("1", "ana") // stores "user:1" in mem
type Prefixed struct {
	Prefix string
	Inner  Store
}

// Get looks up Prefix+key in the inner store.
func (p Prefixed) Get(key string) (string, bool) {
	// TODO: implement me.
	return "", false
}

// Set stores value under Prefix+key in the inner store.
func (p Prefixed) Set(key, value string) {
	// TODO: implement me.
}

// Counting wraps another Store and counts calls to Get and Set.
// A Get counts as a Hit when the key was present and a Miss otherwise.
type Counting struct {
	Inner              Store
	Hits, Misses, Sets int
}

// Get forwards to Inner and records a hit or a miss.
func (c *Counting) Get(key string) (string, bool) {
	// TODO: implement me.
	return "", false
}

// Set forwards to Inner and counts the call.
func (c *Counting) Set(key, value string) {
	// TODO: implement me.
}
