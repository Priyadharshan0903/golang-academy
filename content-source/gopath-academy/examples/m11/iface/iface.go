// Package iface holds the examples for lesson 11.5.
package iface

import (
	"math"
	"unsafe"
)

// [start types]

type Shape interface{ Area() float64 }

type Square struct{ Side float64 }

func (s Square) Area() float64 { return s.Side * s.Side }

type Circle struct{ R float64 }

func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

// [end types]

// [start words]

// eface mirrors how the runtime lays out an empty interface (any): a
// pointer to the dynamic type, and a pointer to the value. This is for
// learning only. Real code must never depend on it.
type eface struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

func words(v any) eface { return *(*eface)(unsafe.Pointer(&v)) }

// [end words]

// [start sum]

// SumDirect calls Area on a concrete type: a static call the compiler
// can inline.
func SumDirect(xs []Square) float64 {
	total := 0.0
	for _, x := range xs {
		total += x.Area()
	}
	return total
}

// SumInterface calls Area through an interface: the method is looked
// up in the itab at run time, and cannot be inlined.
func SumInterface(xs []Shape) float64 {
	total := 0.0
	for _, x := range xs {
		total += x.Area()
	}
	return total
}

// SumGeneric is generic over the element type.
func SumGeneric[T Shape](xs []T) float64 {
	total := 0.0
	for _, x := range xs {
		total += x.Area()
	}
	return total
}

// [end sum]
