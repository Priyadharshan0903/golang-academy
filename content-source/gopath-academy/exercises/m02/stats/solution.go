//go:build solution

// Package stats practices multiple return values, variadic
// functions and closures.
package stats

// Sum returns the total of all the numbers passed in.
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// MinMax returns the smallest and largest of nums.
func MinMax(nums ...int) (lo, hi int, ok bool) {
	if len(nums) == 0 {
		return 0, 0, false
	}
	lo, hi = nums[0], nums[0]
	for _, n := range nums[1:] {
		lo = min(lo, n) // min and max are built-in since Go 1.21
		hi = max(hi, n)
	}
	return lo, hi, true
}

// RunningAverage returns a function that keeps a running average.
func RunningAverage() func(float64) float64 {
	var total float64
	var count int
	return func(v float64) float64 {
		total += v
		count++
		return total / float64(count)
	}
}
