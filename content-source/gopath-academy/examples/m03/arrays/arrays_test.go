package arrays

import "fmt"

// [start basics]

func Example_basics() {
	var scores [3]int // length is part of the type: [3]int
	scores[0] = 90
	fmt.Println(scores, len(scores))

	primes := [...]int{2, 3, 5, 7} // ... lets the compiler count
	fmt.Println(len(primes))

	// [3]int and [4]int are different types:
	// scores = primes would not compile.

	// Output:
	// [90 0 0] 3
	// 4
}

// [end basics]

// [start value]

func Example_valueSemantics() {
	a := [3]string{"x", "y", "z"}
	b := a // copies all three elements
	b[0] = "CHANGED"
	fmt.Println(a[0], b[0])

	// Arrays of comparable elements can be compared with ==.
	fmt.Println(a == [3]string{"x", "y", "z"})

	// Output:
	// x CHANGED
	// true
}

// [end value]

// [start key]

// Because arrays are comparable values, they make good map keys.
type point [2]int

func Example_asMapKey() {
	visited := map[point]bool{}
	visited[point{1, 2}] = true
	fmt.Println(visited[point{1, 2}], visited[point{2, 1}])

	// Output:
	// true false
}

// [end key]
