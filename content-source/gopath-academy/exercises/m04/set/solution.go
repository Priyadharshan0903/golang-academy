//go:build solution

// Package set practices generics: a generic type, generic methods
// on it, and functions whose constraints are stricter than the type's.
package set

import (
	"cmp"
	"maps"
	"slices"
)

// Set is an unordered collection of unique values.
type Set[T comparable] struct {
	m map[T]struct{}
}

// New returns a set containing items.
func New[T comparable](items ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{}, len(items))}
	for _, v := range items {
		s.Add(v)
	}
	return s
}

// Add inserts v.
func (s *Set[T]) Add(v T) { s.m[v] = struct{}{} }

// Has reports whether v is in the set.
func (s *Set[T]) Has(v T) bool {
	_, ok := s.m[v]
	return ok
}

// Len returns the number of values in the set.
func (s *Set[T]) Len() int { return len(s.m) }

// Union returns a new set with every value from s and other.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	out := &Set[T]{m: maps.Clone(s.m)}
	for v := range other.m {
		out.Add(v)
	}
	return out
}

// Sorted returns the values of s in ascending order.
func Sorted[T cmp.Ordered](s *Set[T]) []T {
	return slices.Sorted(maps.Keys(s.m))
}
