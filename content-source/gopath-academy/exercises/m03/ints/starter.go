//go:build !solution

// Package ints practices slices: length, capacity, append and copying.
package ints

// Chunk splits s into consecutive pieces of at most size elements.
//
//	Chunk([]int{1, 2, 3, 4, 5}, 2) == [][]int{{1, 2}, {3, 4}, {5}}
//
// Chunk returns nil for an empty s. If size < 1 it panics.
// The chunks must not share memory with s: appending to a chunk
// must never change s or another chunk.
func Chunk(s []int, size int) [][]int {
	// TODO: implement me.
	return nil
}

// RemoveAt returns s without the element at index i, keeping order.
// It must not modify the caller's slice.
func RemoveAt(s []int, i int) []int {
	// TODO: implement me.
	return nil
}

// Dedupe returns the values of s in first-seen order with duplicates
// removed. Dedupe([]int{3, 1, 3, 2, 1}) == []int{3, 1, 2}.
// The result must not share memory with s.
func Dedupe(s []int) []int {
	// TODO: implement me.
	return nil
}
