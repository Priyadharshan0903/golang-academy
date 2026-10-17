//go:build !solution

// Package set practices generics: a generic type, generic methods
// on it, and functions whose constraints are stricter than the type's.
package set

import "cmp"

// Set is an unordered collection of unique values.
// The zero value is NOT ready to use; create one with New.
type Set[T comparable] struct {
	// TODO: add the fields you need.
}

// New returns a set containing items.
func New[T comparable](items ...T) *Set[T] {
	// TODO: implement me.
	return &Set[T]{}
}

// Add inserts v. Adding a value that is already present does nothing.
func (s *Set[T]) Add(v T) {
	// TODO: implement me.
}

// Has reports whether v is in the set.
func (s *Set[T]) Has(v T) bool {
	// TODO: implement me.
	return false
}

// Len returns the number of values in the set.
func (s *Set[T]) Len() int {
	// TODO: implement me.
	return 0
}

// Union returns a NEW set with every value from s and other.
// Neither s nor other is modified.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	// TODO: implement me.
	return nil
}

// Sorted returns the values of s in ascending order. It is a function,
// not a method, because it needs a stricter constraint (cmp.Ordered)
// than the Set type itself (comparable).
func Sorted[T cmp.Ordered](s *Set[T]) []T {
	// TODO: implement me.
	return nil
}
