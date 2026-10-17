// Package notes practices building a JSON API with http.ServeMux.
package notes

import (
	"slices"
	"sync"
)

// Note is one note. The JSON field names are part of the API.
type Note struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Store is a concurrency-safe, in-memory set of notes. It is complete;
// the exercise is the HTTP layer in front of it.
type Store struct {
	mu     sync.Mutex
	notes  map[int]Note
	nextID int
}

// NewStore returns an empty store.
func NewStore() *Store { return &Store{notes: map[int]Note{}, nextID: 1} }

// Create stores n with a new ID and returns it.
func (s *Store) Create(n Note) Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	n.ID = s.nextID
	s.nextID++
	s.notes[n.ID] = n
	return n
}

// Get returns the note with the given ID.
func (s *Store) Get(id int) (Note, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.notes[id]
	return n, ok
}

// List returns all notes, ordered by ID.
func (s *Store) List() []Note {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Note, 0, len(s.notes))
	for _, n := range s.notes {
		out = append(out, n)
	}
	slices.SortFunc(out, func(a, b Note) int { return a.ID - b.ID })
	return out
}

// Delete removes the note with the given ID and reports whether it existed.
func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.notes[id]
	delete(s.notes, id)
	return ok
}
