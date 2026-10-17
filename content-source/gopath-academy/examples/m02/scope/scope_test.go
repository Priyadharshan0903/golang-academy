package scope

import (
	"errors"
	"fmt"
)

func load(name string) (string, error) {
	if name == "" {
		return "", errors.New("empty name")
	}
	return "data for " + name, nil
}

// [start blocks]

func Example_blocks() {
	x := "outer"
	{
		x := "inner" // a NEW variable that shadows the outer x
		fmt.Println(x)
	}
	fmt.Println(x)

	if y := len(x); y > 3 {
		fmt.Println("y exists only in this if:", y)
	}
	// fmt.Println(y) would not compile here

	// Output:
	// inner
	// outer
	// y exists only in this if: 5
}

// [end blocks]

// [start errshadow]

// buggy intends to report the error from load, but := inside the if
// block declares a brand new err that disappears at the closing brace.
func buggy(name string) error {
	var err error
	if name != "skip" {
		data, err := load(name) // shadows the outer err
		_ = data
		_ = err
	}
	return err // always nil
}

// fixed assigns to the existing variable with = instead.
func fixed(name string) error {
	var err error
	if name != "skip" {
		_, err = load(name)
	}
	return err
}

func Example_shadowedError() {
	fmt.Println(buggy(""))
	fmt.Println(fixed(""))

	// Output:
	// <nil>
	// empty name
}

// [end errshadow]

// [start universe]

func Example_shadowBuiltins() {
	// Built-in names like len, true or string are not keywords.
	// They live in the "universe" scope and CAN be shadowed.
	len := 3 // now len is an int in this function
	fmt.Println(len)
	// len("abc") would fail to compile here: len is not a function

	// Output:
	// 3
}

// [end universe]
