package examplesdoc_test

import (
	"fmt"
	"maps"

	"github.com/Priyadharshan0903/gopath-academy/examples/m09/examplesdoc"
)

// [start names]

// Example is the package example: shown at the top of the package docs.
func Example() {
	var s examplesdoc.Stack
	s.Push("a")
	s.Push("b")
	v, _ := s.Pop()
	fmt.Println(v, s.Len())
	// Output: b 1
}

// ExampleStack documents the type Stack.
func ExampleStack() {
	var s examplesdoc.Stack
	fmt.Println(s.Len())
	// Output: 0
}

// ExampleStack_Pop documents the method Stack.Pop.
func ExampleStack_Pop() {
	var s examplesdoc.Stack
	_, ok := s.Pop()
	fmt.Println(ok)
	// Output: false
}

// ExampleStack_Pop_order is a second example for Pop; the suffix
// after the method name starts with a lowercase letter.
func ExampleStack_Pop_order() {
	var s examplesdoc.Stack
	for _, v := range []string{"x", "y", "z"} {
		s.Push(v)
	}
	for s.Len() > 0 {
		v, _ := s.Pop()
		fmt.Print(v)
	}
	fmt.Println()
	// Output: zyx
}

// [end names]

// [start unordered]

func Example_unordered() {
	set := map[string]bool{"go": true, "rust": true, "zig": true}
	for k := range maps.Keys(set) { // random order
		fmt.Println(k)
	}
	// Unordered output:
	// rust
	// go
	// zig
}

// [end unordered]
