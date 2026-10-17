package functions

import (
	"errors"
	"fmt"
	"strings"
)

// [start multi]

// divide returns the quotient, or an error when b is zero.
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func Example_multipleReturns() {
	if q, err := divide(10, 4); err == nil {
		fmt.Println(q)
	}
	_, err := divide(1, 0)
	fmt.Println(err)

	// Output:
	// 2.5
	// division by zero
}

// [end multi]

// [start named]

// split uses named results. They start at their zero values,
// and a bare "return" returns their current values.
func split(full string) (first, last string) {
	first, last, _ = strings.Cut(full, " ")
	return
}

func Example_namedResults() {
	f, l := split("Rob Pike")
	fmt.Println(l + ", " + f)

	// Output:
	// Pike, Rob
}

// [end named]

// [start variadic]

func sum(nums ...int) int { // nums is a []int inside the function
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func Example_variadic() {
	fmt.Println(sum())
	fmt.Println(sum(1, 2, 3))

	values := []int{10, 20, 30}
	fmt.Println(sum(values...)) // spread an existing slice

	// Output:
	// 0
	// 6
	// 60
}

// [end variadic]

// [start values]

func apply(nums []int, f func(int) int) []int {
	out := make([]int, 0, len(nums))
	for _, n := range nums {
		out = append(out, f(n))
	}
	return out
}

func Example_functionValues() {
	double := func(n int) int { return n * 2 }

	fmt.Println(apply([]int{1, 2, 3}, double))
	fmt.Println(apply([]int{1, 2, 3}, func(n int) int { return n * n }))

	// Output:
	// [2 4 6]
	// [1 4 9]
}

// [end values]

// [start closures]

// counter returns a function that remembers count between calls.
func counter() func() int {
	count := 0
	return func() int {
		count++ // captures the variable, not a copy of its value
		return count
	}
}

func Example_closures() {
	next := counter()
	fmt.Println(next(), next(), next())

	other := counter() // each call creates a new, independent count
	fmt.Println(other())

	// Output:
	// 1 2 3
	// 1
}

// [end closures]

// [start loopvar]

func Example_loopVariables() {
	var funcs []func() int
	for i := range 3 {
		// Since Go 1.22 every iteration has its own i, so each
		// closure captures a different variable.
		funcs = append(funcs, func() int { return i })
	}
	for _, f := range funcs {
		fmt.Println(f())
	}

	// Output:
	// 0
	// 1
	// 2
}

// [end loopvar]
