package controlflow

import (
	"fmt"
	"strconv"
)

// [start if]

func Example_ifWithInit() {
	// The init statement's variables exist only inside the if/else.
	if n, err := strconv.Atoi("42"); err != nil {
		fmt.Println("not a number:", err)
	} else if n > 40 {
		fmt.Println("big:", n)
	} else {
		fmt.Println("small:", n)
	}

	// Output:
	// big: 42
}

// [end if]

// [start for]

func Example_forLoops() {
	// 1. Classic three-part loop.
	for i := 0; i < 3; i++ {
		fmt.Println("classic", i)
	}

	// 2. Condition only: Go's "while".
	n := 1
	for n < 100 {
		n *= 3
	}
	fmt.Println(n)

	// 3. Range over an integer (Go 1.22+): i goes 0, 1.
	for i := range 2 {
		fmt.Println("range int", i)
	}

	// 4. Range over a slice gives index and value.
	for i, name := range []string{"ana", "bo"} {
		fmt.Println(i, name)
	}

	// 5. Infinite loop with break.
	tries := 0
	for {
		tries++
		if tries == 3 {
			break
		}
	}
	fmt.Println("tries:", tries)

	// Output:
	// classic 0
	// classic 1
	// classic 2
	// 243
	// range int 0
	// range int 1
	// 0 ana
	// 1 bo
	// tries: 3
}

// [end for]

// [start switch]

func describe(code int) string {
	switch code {
	case 200, 201, 204: // several values in one case
		return "success"
	case 301, 302:
		return "redirect"
	case 404:
		return "not found"
	default:
		return "other"
	}
	// No break needed: Go never falls through unless you say so.
}

func grade(score int) string {
	switch { // no expression: same as "switch true"
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	default:
		return "C"
	}
}

func Example_switch() {
	fmt.Println(describe(204), describe(302), describe(500))
	fmt.Println(grade(95), grade(80), grade(10))

	// Output:
	// success redirect other
	// A B C
}

// [end switch]

// [start fallthrough]

func Example_fallthrough() {
	switch n := 1; n {
	case 1:
		fmt.Println("one")
		fallthrough // run the next case body, without checking it
	case 2:
		fmt.Println("two")
	case 3:
		fmt.Println("three")
	}

	// Output:
	// one
	// two
}

// [end fallthrough]

// [start labels]

func Example_labels() {
	grid := [][]int{
		{1, 2, 3},
		{4, -1, 6},
		{7, 8, 9},
	}

search:
	for r, row := range grid {
		for c, v := range row {
			if v < 0 {
				fmt.Printf("negative at row %d, col %d\n", r, c)
				break search // leaves BOTH loops
			}
		}
	}

	// Output:
	// negative at row 1, col 1
}

// [end labels]

// [start breakswitch]

func Example_breakInsideSwitch() {
	// break inside a switch only leaves the switch, not the loop.
	for i := range 5 {
		switch i {
		case 2:
			break // ends the switch; the loop keeps going
		}
		fmt.Println("unlabeled", i)
		if i == 2 {
			break // this one is outside the switch: it ends the loop
		}
	}

loop:
	for i := range 5 {
		switch i {
		case 1:
			break loop // a label makes break leave the loop
		}
		fmt.Println("labeled", i)
	}

	// Output:
	// unlabeled 0
	// unlabeled 1
	// unlabeled 2
	// labeled 0
}

// [end breakswitch]
