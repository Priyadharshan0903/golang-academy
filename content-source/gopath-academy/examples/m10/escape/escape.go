// Package escape holds the examples for lesson 10.4.
package escape

// [start code]

type Point struct{ X, Y, Z float64 }

// The //go:noinline directives keep the compiler from inlining these
// tiny functions, so the examples show each function's own escape
// behavior. In real code, let the compiler inline.

//go:noinline
func NewPointValue(x float64) Point { return Point{x, x, x} } // returned by value: no heap

//go:noinline
func NewPointPtr(x float64) *Point { return &Point{x, x, x} } // escapes: outlives the call

// lastSeen makes values passed to Remember escape to the heap.
var lastSeen any

//go:noinline
func Remember(v any) { lastSeen = v } // v escapes: it is stored in a global

//go:noinline
func IsInt(v any) bool { // v does not escape: only inspected
	_, ok := v.(int)
	return ok
}

//go:noinline
func SumFixed() int {
	var buf [64]int // fixed-size array: lives on the stack
	for i := range buf {
		buf[i] = i
	}
	total := 0
	for _, v := range buf {
		total += v
	}
	return total
}

// [end code]
