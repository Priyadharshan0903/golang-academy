package generics

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"
)

// [start func]

// Map works for any input type T and output type U.
func Map[T, U any](items []T, f func(T) U) []U {
	out := make([]U, 0, len(items))
	for _, v := range items {
		out = append(out, f(v))
	}
	return out
}

func Example_genericFunction() {
	lengths := Map([]string{"go", "rust", "zig"}, func(s string) int { return len(s) })
	fmt.Println(lengths)

	labels := Map(lengths, strconv.Itoa) // T and U are inferred from the arguments
	fmt.Printf("%q\n", labels)

	explicit := Map[int, int]([]int{1, 2}, func(n int) int { return n * 10 })
	fmt.Println(explicit)

	// Output:
	// [2 4 3]
	// ["2" "4" "3"]
	// [10 20]
}

// [end func]

// [start constraints]

// MaxOf needs values it can compare with >, so it constrains T to
// cmp.Ordered (all integer, float and string types).
func MaxOf[T cmp.Ordered](first T, rest ...T) T {
	m := first
	for _, v := range rest {
		if v > m {
			m = v
		}
	}
	return m
}

// IndexOf only needs ==, so comparable is enough.
func IndexOf[T comparable](items []T, target T) int {
	for i, v := range items {
		if v == target {
			return i
		}
	}
	return -1
}

func Example_constraints() {
	fmt.Println(MaxOf(3, 9, 2), MaxOf("pear", "apple"), MaxOf(1.5, -2.0))
	fmt.Println(IndexOf([]string{"a", "b", "c"}, "c"))

	// Output:
	// 9 pear 1.5
	// 2
}

// [end constraints]

// [start union]

// Number is a constraint made of a type union. The ~ means "any type
// whose underlying type is", so named types like Celsius qualify too.
type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](nums ...T) T {
	var total T // the zero value of whatever T is
	for _, n := range nums {
		total += n
	}
	return total
}

type Celsius float64

func Example_typeUnions() {
	fmt.Println(Sum(1, 2, 3))
	fmt.Println(Sum(Celsius(20.5), Celsius(1.5)))

	// Output:
	// 6
	// 22
}

// [end union]

// [start type]

// Stack is a generic type: the element type is a parameter.
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

// Pop returns the top element, or the zero value and false if empty.
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

func (s *Stack[T]) Len() int { return len(s.items) }

func Example_genericType() {
	var words Stack[string]
	words.Push("first")
	words.Push("second")
	top, _ := words.Pop()
	fmt.Println(top, words.Len())

	var nums Stack[int]
	_, ok := nums.Pop()
	fmt.Println(ok)

	// Output:
	// second 1
	// false
}

// [end type]

// [start methodconstraint]

// A constraint can require methods too, not just types.
type Named interface {
	Name() string
}

type User struct{ name string }

func (u User) Name() string { return u.name }

func JoinNames[T Named](items []T) string {
	names := Map(items, func(v T) string { return v.Name() })
	return strings.Join(names, ", ")
}

func Example_methodConstraint() {
	fmt.Println(JoinNames([]User{{"ana"}, {"bo"}}))

	// Output:
	// ana, bo
}

// [end methodconstraint]
