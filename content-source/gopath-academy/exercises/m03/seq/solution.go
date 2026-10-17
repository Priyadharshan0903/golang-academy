//go:build solution

// Package seq practices writing and composing iterators.
package seq

import "iter"

// Range yields start, start+step, ... while the value is below end.
func Range(start, end, step int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for v := start; v < end; v += step {
			if !yield(v) {
				return
			}
		}
	}
}

// Map yields f(v) for every v produced by s.
func Map(s iter.Seq[int], f func(int) int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for v := range s {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Take yields at most the first n values of s.
func Take(s iter.Seq[int], n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		if n <= 0 {
			return
		}
		count := 0
		for v := range s {
			if !yield(v) {
				return
			}
			count++
			if count == n {
				return
			}
		}
	}
}
