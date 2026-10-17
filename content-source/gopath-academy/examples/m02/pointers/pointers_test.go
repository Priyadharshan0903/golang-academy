package pointers

import "fmt"

// [start basics]

func Example_basics() {
	x := 10
	p := &x // p has type *int and holds the address of x

	fmt.Println(*p) // dereference: read the value p points to
	*p = 20         // write through the pointer
	fmt.Println(x)

	var nothing *int // the zero value of a pointer is nil
	fmt.Println(nothing == nil)

	// Output:
	// 10
	// 20
	// true
}

// [end basics]

// [start passing]

// Go always passes arguments by value: the function gets a copy.
func resetValue(n int) { n = 0 }

// Passing a pointer copies the address, so the function can
// change the caller's variable.
func resetPointer(n *int) { *n = 0 }

func Example_passByValue() {
	a, b := 5, 5
	resetValue(a)
	resetPointer(&b)
	fmt.Println(a, b)

	// Output:
	// 5 0
}

// [end passing]

// [start structs]

type Player struct {
	Name  string
	Score int
}

func addPoints(p *Player, points int) {
	p.Score += points // same as (*p).Score: Go dereferences for you
}

func Example_structPointers() {
	ana := Player{Name: "Ana"}
	addPoints(&ana, 10)
	addPoints(&ana, 5)
	fmt.Println(ana.Name, ana.Score)

	// Output:
	// Ana 15
}

// [end structs]

// [start escape]

// newPlayer returns a pointer to a local variable. This is safe in
// Go: the compiler notices the variable outlives the function and
// allocates it on the heap ("escape analysis").
func newPlayer(name string) *Player {
	p := Player{Name: name}
	return &p
}

func Example_returnPointer() {
	bo := newPlayer("Bo")
	bo.Score = 7
	fmt.Println(bo.Name, bo.Score)

	n := new(int) // new(T) allocates a zero T and returns *T
	*n = 3
	fmt.Println(*n)

	// Output:
	// Bo 7
	// 3
}

// [end escape]

// [start nil]

func scoreOf(p *Player) int {
	if p == nil { // guard before dereferencing
		return 0
	}
	return p.Score
}

func Example_nilPointer() {
	var missing *Player
	fmt.Println(scoreOf(missing))

	defer func() { fmt.Println("recovered:", recover()) }()
	fmt.Println(missing.Score) // dereferencing nil panics

	// Output:
	// 0
	// recovered: runtime error: invalid memory address or nil pointer dereference
}

// [end nil]
