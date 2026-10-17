package slicegotchas

import (
	"fmt"
	"slices"
)

// [start alias]

func Example_appendAliasing() {
	base := make([]int, 3, 10) // spare capacity
	a := append(base, 1)       // writes into base's array at index 3
	b := append(base, 2)       // ALSO writes at index 3

	fmt.Println(a[3], b[3]) // a was overwritten by the second append

	// Output:
	// 2 2
}

// [end alias]

// [start clone]

func Example_avoidAliasing() {
	base := make([]int, 3, 10)
	a := append(slices.Clone(base), 1) // independent copy first
	b := append(slices.Clone(base), 2)
	fmt.Println(a[3], b[3])

	// The full slice expression s[low:high:max] caps capacity,
	// forcing the next append to allocate a new array.
	c := append(base[:3:3], 3)
	d := append(base[:3:3], 4)
	fmt.Println(c[3], d[3])

	// Output:
	// 1 2
	// 3 4
}

// [end clone]

// [start func]

// addBroken appends, but the caller never sees the new length,
// because the slice header was passed by value.
func addBroken(s []string, v string) {
	s = append(s, v)
}

// add returns the updated slice, like append itself does.
func add(s []string, v string) []string {
	return append(s, v)
}

func Example_appendInFunction() {
	var names []string
	addBroken(names, "ana")
	fmt.Println(len(names))

	names = add(names, "ana")
	fmt.Println(len(names))

	// Output:
	// 0
	// 1
}

// [end func]

// [start modify]

func double(s []int) {
	for i := range s {
		s[i] *= 2 // modifies the caller's elements: same backing array
	}
}

func Example_modifyElements() {
	nums := []int{1, 2, 3}
	double(nums)
	fmt.Println(nums)

	// Output:
	// [2 4 6]
}

// [end modify]

// [start retain]

// firstBytes keeps only a small prefix of a big buffer. Returning
// big[:n] would keep the entire big array alive in memory.
func firstBytes(big []byte, n int) []byte {
	return slices.Clone(big[:n])
}

func Example_memoryRetention() {
	big := make([]byte, 1<<20) // 1 MiB
	header := firstBytes(big, 4)
	fmt.Println(len(header), cap(header) < 1<<20)

	// Output:
	// 4 true
}

// [end retain]

// [start delete]

func Example_deleteWhileFiltering() {
	nums := []int{1, 2, 3, 4, 5, 6}

	// Filter in place: keep reuses nums' backing array.
	keep := nums[:0]
	for _, n := range nums {
		if n%2 == 0 {
			keep = append(keep, n)
		}
	}
	fmt.Println(keep)

	// Or let the standard library do it.
	odd := slices.DeleteFunc([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 })
	fmt.Println(odd)

	// Output:
	// [2 4 6]
	// [1 3 5]
}

// [end delete]
