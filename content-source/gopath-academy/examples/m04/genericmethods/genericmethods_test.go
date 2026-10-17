//go:build go1.27

package genericmethods

import (
	"fmt"
	"strconv"
)

// [start method]

type List[T any] struct {
	items []T
}

func (l *List[T]) Push(v T) { l.items = append(l.items, v) }

// Map is a generic method (Go 1.27+): it declares its own type
// parameter U, in addition to the type's parameter T.
func (l *List[T]) Map[U any](f func(T) U) *List[U] {
	out := &List[U]{}
	for _, v := range l.items {
		out.Push(f(v))
	}
	return out
}

func Example_genericMethod() {
	nums := &List[int]{}
	nums.Push(1)
	nums.Push(2)

	strs := nums.Map(strconv.Itoa) // U inferred as string
	fmt.Printf("%q\n", strs.items)

	tens := nums.Map[int](func(n int) int { return n * 10 }) // explicit U
	fmt.Println(tens.items)

	// Output:
	// ["1" "2"]
	// [10 20]
}

// [end method]

// [start nongeneric]

// Generic methods also work on non-generic types.
type Celsius float64

func (c Celsius) As[T ~float64 | ~float32](convert func(float64) float64) T {
	return T(convert(float64(c)))
}

type Fahrenheit float64

func Example_methodOnPlainType() {
	f := Celsius(100).As[Fahrenheit](func(c float64) float64 { return c*9/5 + 32 })
	fmt.Printf("%T %v\n", f, f)

	// Output:
	// genericmethods.Fahrenheit 212
}

// [end nongeneric]
