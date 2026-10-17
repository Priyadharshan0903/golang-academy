package link

import (
	"context"
	"sync"
)

// MemoryStore is an in-memory Store, safe for concurrent use.
// It is used in tests and for local development.
type MemoryStore struct {
	mu    sync.RWMutex // locking is covered in Module 6
	links map[string]Link
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{links: make(map[string]Link)}
}

// Create stores l, or returns ErrCodeTaken if its code exists.
func (m *MemoryStore) Create(_ context.Context, l Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.links[l.Code]; exists {
		return ErrCodeTaken
	}
	m.links[l.Code] = l
	return nil
}

// Get returns the link for code, or ErrNotFound.
func (m *MemoryStore) Get(_ context.Context, code string) (Link, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return l, nil
}

// Compile-time check that MemoryStore satisfies Store.
var _ Store = (*MemoryStore)(nil)

// Delete removes the link for code, or returns ErrNotFound.
func (m *MemoryStore) Delete(_ context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.links[code]; !ok {
		return ErrNotFound
	}
	delete(m.links, code)
	return nil
}
