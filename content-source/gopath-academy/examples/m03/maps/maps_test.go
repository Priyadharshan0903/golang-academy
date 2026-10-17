package maps

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// [start basics]

func Example_basics() {
	ages := map[string]int{ // map literal
		"ana": 25,
		"bo":  41,
	}
	ages["cara"] = 30 // insert or update
	delete(ages, "bo")
	fmt.Println(len(ages), ages["ana"])

	// A missing key returns the zero value, not an error.
	fmt.Println(ages["nobody"])

	// The comma-ok form tells "missing" apart from "zero".
	if age, ok := ages["cara"]; ok {
		fmt.Println("cara is", age)
	}
	_, ok := ages["nobody"]
	fmt.Println(ok)

	// Output:
	// 2 25
	// 0
	// cara is 30
	// false
}

// [end basics]

// [start nil]

func Example_nilMap() {
	var m map[string]int // nil map

	fmt.Println(m["x"], len(m)) // reading a nil map is fine

	defer func() { fmt.Println("recovered:", recover()) }()
	m["x"] = 1 // writing panics: use make or a literal first

	// Output:
	// 0 0
	// recovered: assignment to entry in nil map
}

// [end nil]

// [start count]

func Example_counting() {
	words := strings.Fields("the cat and the hat and the bat")

	counts := make(map[string]int)
	for _, w := range words {
		counts[w]++ // the zero value makes counting a one-liner
	}

	for _, w := range slices.Sorted(maps.Keys(counts)) {
		fmt.Println(w, counts[w])
	}

	// Output:
	// and 2
	// bat 1
	// cat 1
	// hat 1
	// the 3
}

// [end count]

// [start set]

func Example_set() {
	// map[T]struct{} is the idiomatic set: struct{} takes no memory.
	seen := make(map[string]struct{})
	for _, tag := range []string{"go", "k8s", "go", "infra"} {
		seen[tag] = struct{}{}
	}
	_, hasGo := seen["go"]
	fmt.Println(len(seen), hasGo)

	// Output:
	// 3 true
}

// [end set]

// [start group]

func Example_groupBy() {
	names := []string{"ana", "bo", "anu", "bala", "cara"}

	byLetter := make(map[byte][]string)
	for _, n := range names {
		byLetter[n[0]] = append(byLetter[n[0]], n) // append to nil is fine
	}
	fmt.Println(byLetter['a'], byLetter['b'], byLetter['c'])

	// Output:
	// [ana anu] [bo bala] [cara]
}

// [end group]

// [start structvalues]

type account struct{ balance int }

func Example_structValues() {
	accounts := map[string]account{"ana": {balance: 10}}

	// accounts["ana"].balance += 5 does not compile: map elements
	// are not addressable. Read, modify, write back:
	acc := accounts["ana"]
	acc.balance += 5
	accounts["ana"] = acc

	// Or store pointers when values are updated often.
	ptrs := map[string]*account{"bo": {balance: 1}}
	ptrs["bo"].balance += 5

	fmt.Println(accounts["ana"].balance, ptrs["bo"].balance)

	// Output:
	// 15 6
}

// [end structvalues]
