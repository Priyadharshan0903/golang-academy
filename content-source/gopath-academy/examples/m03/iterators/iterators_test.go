package iterators

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"strings"
)

// [start seq]

// Countdown returns an iterator: a function that calls yield once
// per value. iter.Seq[int] is defined as func(yield func(int) bool).
func Countdown(from int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := from; i >= 0; i-- {
			if !yield(i) { // the loop body asked to stop (break)
				return
			}
		}
	}
}

func Example_rangeOverFunc() {
	for n := range Countdown(3) {
		fmt.Print(n, ";")
	}
	fmt.Println()

	for n := range Countdown(100) {
		if n < 98 {
			break // yield returns false, Countdown returns early
		}
		fmt.Println(n)
	}

	// Output:
	// 3;2;1;0;
	// 100
	// 99
	// 98
}

// [end seq]

// [start seq2]

// Enumerate yields index and value pairs, like range over a slice.
func Enumerate(words []string) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i, w := range words {
			if !yield(i, w) {
				return
			}
		}
	}
}

func Example_seq2() {
	for i, w := range Enumerate([]string{"go", "is", "fun"}) {
		fmt.Println(i, w)
	}

	// Output:
	// 0 go
	// 1 is
	// 2 fun
}

// [end seq2]

// [start compose]

// Filter wraps an iterator and only passes on values that match.
// Nothing runs until someone ranges over the result: iterators are lazy.
func Filter(seq iter.Seq[int], keep func(int) bool) iter.Seq[int] {
	return func(yield func(int) bool) {
		for v := range seq {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

func Example_composition() {
	even := Filter(Countdown(10), func(n int) bool { return n%2 == 0 })
	fmt.Println(slices.Collect(even)) // gather an iterator into a slice

	// Output:
	// [10 8 6 4 2 0]
}

// [end compose]

// [start stdlib]

func Example_standardIterators() {
	names := []string{"bo", "ana", "cara"}

	for i, n := range slices.All(names) { // same as range names
		fmt.Print(i, "=", n, ";")
	}
	fmt.Println()

	upper := slices.Collect(func(yield func(string) bool) {
		for n := range slices.Values(names) {
			if !yield(strings.ToUpper(n)) {
				return
			}
		}
	})
	fmt.Println(upper)

	ages := map[string]int{"bo": 41, "ana": 25}
	fmt.Println(slices.Sorted(maps.Keys(ages)))

	for line := range strings.Lines("a\nb\n") { // Go 1.24+
		fmt.Printf("%q;", line)
	}
	fmt.Println()

	// Output:
	// 0=bo;1=ana;2=cara;
	// [BO ANA CARA]
	// [ana bo]
	// "a\n";"b\n";
}

// [end stdlib]

// [start pull]

func Example_pull() {
	// iter.Pull turns a push iterator into next/stop functions,
	// useful when you need to walk two sequences side by side.
	next, stop := iter.Pull(Countdown(2))
	defer stop() // always call stop to release resources

	for {
		v, ok := next()
		if !ok {
			break
		}
		fmt.Println("pulled", v)
	}

	// Output:
	// pulled 2
	// pulled 1
	// pulled 0
}

// [end pull]
