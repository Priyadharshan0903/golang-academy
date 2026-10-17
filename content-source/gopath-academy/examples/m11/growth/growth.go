// Package growth holds the examples for lesson 11.6.
package growth

// [start caps]

// Kept holds the last slice built by CapSteps. Storing it in a package
// variable makes the slice escape, so its backing array is always on
// the heap.
var Kept []int

// CapSteps appends n ints one at a time and records each new capacity.
func CapSteps(n int) []int {
	var s []int
	var steps []int
	for i := range n {
		s = append(s, i)
		if len(steps) == 0 || steps[len(steps)-1] != cap(s) {
			steps = append(steps, cap(s))
		}
	}
	Kept = s
	return steps
}

// CapStepsLocal is the same, except that s never leaves the function.
func CapStepsLocal(n int) []int {
	var s []int
	var steps []int
	for i := range n {
		s = append(s, i)
		if len(steps) == 0 || steps[len(steps)-1] != cap(s) {
			steps = append(steps, cap(s))
		}
	}
	return steps
}

// [end caps]
