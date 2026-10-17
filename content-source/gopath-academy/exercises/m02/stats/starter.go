//go:build !solution

// Package stats practices multiple return values, variadic
// functions and closures.
package stats

// Sum returns the total of all the numbers passed in.
// Sum() is 0.
func Sum(nums ...int) int {
	// TODO: implement me.
	return 0
}

// MinMax returns the smallest and largest of nums.
// ok is false (and lo and hi are 0) when nums is empty.
func MinMax(nums ...int) (lo, hi int, ok bool) {
	// TODO: implement me.
	return 0, 0, false
}

// RunningAverage returns a function that records a new value each
// time it is called and returns the average of all values so far.
//
//	avg := RunningAverage()
//	avg(10) // 10
//	avg(20) // 15
//	avg(30) // 20
//
// Each call to RunningAverage starts a new, independent average.
func RunningAverage() func(float64) float64 {
	// TODO: implement me.
	return func(float64) float64 { return 0 }
}
