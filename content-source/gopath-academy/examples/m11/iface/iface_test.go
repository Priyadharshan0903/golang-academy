package iface

import (
	"fmt"
	"testing"
	"unsafe"
)

// [start example]

func Example_words() {
	x, y := 42, 7
	a, b, c := words(&x), words(&y), words("hello")

	fmt.Println("same dynamic type (*int):", a.typ == b.typ)
	fmt.Println("different from string:", a.typ != c.typ)
	fmt.Println("data word is the pointer itself:", a.data == unsafe.Pointer(&x))
	// Output:
	// same dynamic type (*int): true
	// different from string: true
	// data word is the pointer itself: true
}

func Example_typedNil() {
	var sq *Square   // a nil pointer
	var s Shape = sq // an interface holding (type *Square, value nil)
	var empty Shape  // an interface holding nothing at all
	fmt.Println(s == nil, empty == nil)
	// Output: false true
}

// [end example]

func shapes(n int) ([]Square, []Shape) {
	sq := make([]Square, n)
	sh := make([]Shape, n)
	for i := range n {
		sq[i] = Square{Side: float64(i % 10)}
		if i%2 == 0 {
			sh[i] = sq[i]
		} else {
			sh[i] = Circle{R: float64(i % 10)}
		}
	}
	return sq, sh
}

// [start bench]

var result float64

func BenchmarkDispatch(b *testing.B) {
	sq, sh := shapes(1024)
	b.Run("call=direct", func(b *testing.B) {
		for b.Loop() {
			result = SumDirect(sq)
		}
	})
	b.Run("call=generic", func(b *testing.B) {
		for b.Loop() {
			result = SumGeneric(sq)
		}
	})
	b.Run("call=interface", func(b *testing.B) {
		for b.Loop() {
			result = SumInterface(sh)
		}
	})
	gsh := make([]Shape, len(sq)) // generic instantiated with an interface type
	for i := range sq {
		gsh[i] = sq[i]
	}
	b.Run("call=generic-iface", func(b *testing.B) {
		for b.Loop() {
			result = SumGeneric(gsh)
		}
	})
}

// [end bench]

func TestSums(t *testing.T) {
	sq, _ := shapes(10)
	if SumDirect(sq) != SumGeneric(sq) {
		t.Fatal("direct and generic disagree")
	}
}
