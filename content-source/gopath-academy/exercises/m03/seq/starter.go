//go:build !solution

// Package seq practices writing and composing iterators.
package seq

import "iter"

// Range yields start, start+step, start+2*step, ... while the value
// is less than end. step is always positive.
//
//	slices.Collect(Range(0, 10, 3)) == []int{0, 3, 6, 9}
func Range(start, end, step int) iter.Seq[int] {
	// TODO: implement me.
	return func(yield func(int) bool) {}
}

// Map yields f(v) for every v produced by s.
func Map(s iter.Seq[int], f func(int) int) iter.Seq[int] {
	// TODO: implement me.
	return func(yield func(int) bool) {}
}

// Take yields at most the first n values of s, and then stops pulling
// from s. It must work even when s is infinite.
func Take(s iter.Seq[int], n int) iter.Seq[int] {
	// TODO: implement me.
	return func(yield func(int) bool) {}
}
