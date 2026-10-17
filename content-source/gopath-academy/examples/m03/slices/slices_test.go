package slices

import "fmt"

// [start basics]

func Example_basics() {
	nums := []int{10, 20, 30} // a slice literal: no length in the type
	nums = append(nums, 40)
	fmt.Println(nums, len(nums))

	buf := make([]byte, 3)       // length 3, all zero
	list := make([]string, 0, 8) // length 0, room for 8 before growing
	fmt.Println(buf, len(list), cap(list))

	// Output:
	// [10 20 30 40] 4
	// [0 0 0] 0 8
}

// [end basics]

// [start header]

func Example_sharedBackingArray() {
	arr := [5]int{1, 2, 3, 4, 5}
	s := arr[1:4] // elements 1, 2, 3 (end index is exclusive)

	fmt.Println(s, len(s), cap(s)) // cap runs to the end of arr

	s[0] = 99 // writes into arr: s is a window onto it
	fmt.Println(arr)

	// Output:
	// [2 3 4] 3 4
	// [1 99 3 4 5]
}

// [end header]

// [start append]

func Example_appendGrowth() {
	s := make([]int, 0, 2)

	s = append(s, 1, 2) // fits: same backing array
	fmt.Println(len(s), cap(s))

	s = append(s, 3) // does not fit: Go allocates a bigger array and copies
	fmt.Println(len(s), cap(s) >= 3)

	// Output:
	// 2 2
	// 3 true
}

// [end append]

// [start nil]

func Example_nilVsEmpty() {
	var a []int  // nil slice
	b := []int{} // empty, non-nil slice

	fmt.Println(a == nil, b == nil)
	fmt.Println(len(a), len(b))

	a = append(a, 1) // appending to nil works fine
	for range a {
		fmt.Println("ranging over a works too")
	}

	// Output:
	// true false
	// 0 0
	// ranging over a works too
}

// [end nil]

// [start copy]

func Example_copy() {
	src := []int{1, 2, 3, 4}
	dst := make([]int, 2)
	n := copy(dst, src) // copies min(len(dst), len(src)) elements
	fmt.Println(n, dst)

	// Output:
	// 2 [1 2]
}

// [end copy]

// [start grid]

func Example_twoDimensional() {
	rows, cols := 2, 3
	grid := make([][]int, rows)
	for i := range grid {
		grid[i] = make([]int, cols) // each row is its own slice
	}
	grid[1][2] = 7
	fmt.Println(grid)

	// Output:
	// [[0 0 0] [0 0 7]]
}

// [end grid]
