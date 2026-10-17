//go:build solution

// Package ints practices slices: length, capacity, append and copying.
package ints

import "slices"

// Chunk splits s into consecutive pieces of at most size elements.
func Chunk(s []int, size int) [][]int {
	if size < 1 {
		panic("ints.Chunk: size must be at least 1")
	}
	var out [][]int
	for start := 0; start < len(s); start += size {
		end := min(start+size, len(s))
		out = append(out, slices.Clone(s[start:end]))
	}
	return out
}

// RemoveAt returns s without the element at index i, keeping order.
func RemoveAt(s []int, i int) []int {
	out := make([]int, 0, len(s)-1)
	out = append(out, s[:i]...)
	return append(out, s[i+1:]...)
}

// Dedupe returns the values of s in first-seen order without duplicates.
func Dedupe(s []int) []int {
	seen := make(map[int]struct{}, len(s))
	var out []int
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
