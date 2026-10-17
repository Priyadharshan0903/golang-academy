//go:build !solution

// Package ptr practices pointers: reading and writing through them,
// and handling nil safely.
package ptr

// Swap exchanges the values that a and b point to.
//
//	x, y := 1, 2
//	Swap(&x, &y) // now x == 2, y == 1
func Swap(a, b *int) {
	// TODO: implement me.
}

// Clamp changes *p so that it lies within [lo, hi].
// If p is nil, Clamp does nothing (and must not panic).
func Clamp(p *int, lo, hi int) {
	// TODO: implement me.
}

// ValueOr returns *p, or def when p is nil.
func ValueOr(p *int, def int) int {
	// TODO: implement me.
	return 0
}

// IntPtr returns a pointer to a new int holding v. Handy for
// optional fields, for example in JSON structs.
func IntPtr(v int) *int {
	// TODO: implement me.
	return nil
}
