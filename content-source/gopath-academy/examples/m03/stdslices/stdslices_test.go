package stdslices

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// [start slices]

func Example_slicesPackage() {
	nums := []int{5, 2, 8, 2, 9}

	fmt.Println(slices.Contains(nums, 8), slices.Index(nums, 2))
	fmt.Println(slices.Min(nums), slices.Max(nums))

	sorted := slices.Clone(nums) // Sort works in place, so copy first
	slices.Sort(sorted)
	fmt.Println(sorted)

	i, found := slices.BinarySearch(sorted, 8) // needs sorted input
	fmt.Println(i, found)

	// Compact removes consecutive duplicates IN PLACE and returns the
	// shorter slice. The original's tail is zeroed (Go 1.22+), so
	// always use the returned slice.
	unique := slices.Compact(sorted)
	fmt.Println(unique, sorted)

	fmt.Println(slices.Equal(nums, []int{5, 2, 8, 2, 9}))

	// Output:
	// true 1
	// 2 9
	// [2 2 5 8 9]
	// 3 true
	// [2 5 8 9] [2 5 8 9 0]
	// true
}

// [end slices]

// [start sortfunc]

type user struct {
	name string
	age  int
}

func Example_sortFunc() {
	users := []user{{"cara", 30}, {"ana", 25}, {"bo", 30}}

	// Sort by age descending, then by name ascending.
	slices.SortFunc(users, func(a, b user) int {
		return cmp.Or(
			cmp.Compare(b.age, a.age),
			strings.Compare(a.name, b.name),
		)
	})
	fmt.Println(users)

	// Output:
	// [{bo 30} {cara 30} {ana 25}]
}

// [end sortfunc]

// [start edit]

func Example_insertAndDelete() {
	s := []string{"a", "b", "e"}
	s = slices.Insert(s, 2, "c", "d")
	fmt.Println(s)

	s = slices.Delete(s, 0, 2) // remove indexes [0, 2)
	fmt.Println(s)

	slices.Reverse(s)
	fmt.Println(s)

	// Output:
	// [a b c d e]
	// [c d e]
	// [e d c]
}

// [end edit]

// [start maps]

func Example_mapsPackage() {
	ages := map[string]int{"cara": 30, "ana": 25, "bo": 41}

	// Map order is random, so sort the keys for stable output.
	for _, name := range slices.Sorted(maps.Keys(ages)) {
		fmt.Println(name, ages[name])
	}

	copyOf := maps.Clone(ages)
	delete(copyOf, "bo")
	fmt.Println(len(ages), len(copyOf), maps.Equal(ages, copyOf))

	clear(copyOf) // built-in since Go 1.21: removes every key
	fmt.Println(len(copyOf))

	// Output:
	// ana 25
	// bo 41
	// cara 30
	// 3 2 false
	// 0
}

// [end maps]
