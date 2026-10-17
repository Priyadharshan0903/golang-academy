//go:build solution

// Package ptr practices pointers: reading and writing through them,
// and handling nil safely.
package ptr

// Swap exchanges the values that a and b point to.
func Swap(a, b *int) {
	*a, *b = *b, *a
}

// Clamp changes *p so that it lies within [lo, hi].
func Clamp(p *int, lo, hi int) {
	if p == nil {
		return
	}
	*p = min(max(*p, lo), hi)
}

// ValueOr returns *p, or def when p is nil.
func ValueOr(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

// IntPtr returns a pointer to a new int holding v.
func IntPtr(v int) *int {
	return &v
}
