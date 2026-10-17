package benchmarks

import (
	"fmt"
	"testing"
)

func input(n int) []int {
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i * 7
	}
	return nums
}

// [start correctness]

func TestJoinVariantsAgree(t *testing.T) {
	for _, n := range []int{0, 1, 2, 100} {
		nums := input(n)
		want := JoinPlus(nums)
		if got := JoinBuilder(nums); got != want {
			t.Errorf("JoinBuilder(%d) differs", n)
		}
		if got := JoinAppend(nums); got != want {
			t.Errorf("JoinAppend(%d) differs", n)
		}
	}
}

// [end correctness]

// [start bench]

// b.Loop (Go 1.24+) runs the body as many times as needed for a stable
// measurement, keeps setup outside the timed region, and prevents the
// compiler from optimizing the call away.
func BenchmarkJoin(b *testing.B) {
	for _, size := range []int{10, 1000} {
		nums := input(size) // setup: not timed
		b.Run(fmt.Sprintf("plus/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				JoinPlus(nums)
			}
		})
		b.Run(fmt.Sprintf("builder/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				JoinBuilder(nums)
			}
		})
		b.Run(fmt.Sprintf("append/%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				JoinAppend(nums)
			}
		})
	}
}

// [end bench]

// [start allocs]

// testing.AllocsPerRun turns an allocation budget into a regular test.
func TestJoinAppendAllocations(t *testing.T) {
	nums := input(1000)
	allocs := testing.AllocsPerRun(100, func() { JoinAppend(nums) })
	if allocs > 2 { // the buffer and the final string
		t.Errorf("JoinAppend made %.0f allocations, want at most 2", allocs)
	}
}

// [end allocs]
