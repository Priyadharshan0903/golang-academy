package methods

import (
	"fmt"
	"strings"
)

// [start basic]

type Rect struct {
	W, H float64
}

// Area has a value receiver: it gets a copy of the Rect.
func (r Rect) Area() float64 {
	return r.W * r.H
}

// Scale has a pointer receiver: it can modify the original.
func (r *Rect) Scale(f float64) {
	r.W *= f
	r.H *= f
}

func Example_receivers() {
	r := Rect{W: 2, H: 3}
	fmt.Println(r.Area())

	r.Scale(2) // Go takes &r for you, because r is addressable
	fmt.Println(r.W, r.H, r.Area())

	// Output:
	// 6
	// 4 6 24
}

// [end basic]

// [start valuebug]

type Counter struct{ n int }

func (c Counter) IncBroken() { c.n++ } // increments a copy
func (c *Counter) Inc()      { c.n++ }

func Example_valueReceiverBug() {
	var c Counter
	c.IncBroken()
	c.IncBroken()
	fmt.Println(c.n)

	c.Inc()
	c.Inc()
	fmt.Println(c.n)

	// Output:
	// 0
	// 2
}

// [end valuebug]

// [start anytype]

// Methods can be declared on any named type defined in the same
// package, not only on structs.
type Path []string

func (p Path) String() string { return "/" + strings.Join(p, "/") }

type Celsius float64

func (c Celsius) Fahrenheit() float64 { return float64(c)*9/5 + 32 }

func Example_methodsOnAnyType() {
	fmt.Println(Path{"usr", "local", "go"}.String())
	fmt.Println(Celsius(37).Fahrenheit())

	// Output:
	// /usr/local/go
	// 98.6
}

// [end anytype]

// [start nilrecv]

type Node struct {
	Val  int
	Next *Node
}

// Sum works on a nil *Node: a pointer receiver may be nil, and the
// method can handle that case itself.
func (n *Node) Sum() int {
	if n == nil {
		return 0
	}
	return n.Val + n.Next.Sum()
}

func Example_nilReceiver() {
	list := &Node{1, &Node{2, &Node{3, nil}}}
	fmt.Println(list.Sum())

	var empty *Node
	fmt.Println(empty.Sum())

	// Output:
	// 6
	// 0
}

// [end nilrecv]

// [start values]

func Example_methodValues() {
	r := Rect{W: 1, H: 5}

	area := r.Area // method value: bound to this r (copied now)
	r.W = 100
	fmt.Println(area())

	scale := (*Rect).Scale // method expression: receiver becomes a parameter
	scale(&r, 0.5)
	fmt.Println(r.W, r.H)

	// Output:
	// 5
	// 50 2.5
}

// [end values]
