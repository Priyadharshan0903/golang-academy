package interfaces

import (
	"fmt"
	"math"
	"strings"
)

// [start implicit]

// Shape is satisfied by any type that has these two methods.
type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rect struct{ W, H float64 }

func (r Rect) Area() float64      { return r.W * r.H }
func (r Rect) Perimeter() float64 { return 2 * (r.W + r.H) }

type Circle struct{ R float64 }

func (c Circle) Area() float64      { return math.Pi * c.R * c.R }
func (c Circle) Perimeter() float64 { return 2 * math.Pi * c.R }

// No "implements" keyword anywhere: Rect and Circle simply have
// the right methods.

func totalArea(shapes ...Shape) float64 {
	var sum float64
	for _, s := range shapes {
		sum += s.Area()
	}
	return sum
}

func Example_implicitSatisfaction() {
	fmt.Printf("%.2f\n", totalArea(Rect{2, 3}, Circle{1}))

	// Output:
	// 9.14
}

// [end implicit]

// [start small]

// Small interfaces are easy to satisfy and easy to fake in tests.
type Notifier interface {
	Notify(msg string) error
}

type EmailNotifier struct{ To string }

func (e EmailNotifier) Notify(msg string) error {
	fmt.Println("email to", e.To+":", msg)
	return nil
}

// recorder is a test double: no network, just memory.
type recorder struct{ sent []string }

func (r *recorder) Notify(msg string) error {
	r.sent = append(r.sent, msg)
	return nil
}

func alertOnHighLoad(n Notifier, load float64) {
	if load > 0.9 {
		n.Notify(fmt.Sprintf("load is %.0f%%", load*100))
	}
}

func Example_smallInterfaces() {
	alertOnHighLoad(EmailNotifier{To: "oncall@example.com"}, 0.95)

	rec := &recorder{}
	alertOnHighLoad(rec, 0.5)
	alertOnHighLoad(rec, 0.99)
	fmt.Println(rec.sent)

	// Output:
	// email to oncall@example.com: load is 95%
	// [load is 99%]
}

// [end small]

// [start methodset]

type Resetter interface{ Reset() }

type Buffer struct{ data []string }

// Reset has a pointer receiver, so only *Buffer has it in its
// method set. A Buffer value does not satisfy Resetter.
func (b *Buffer) Reset() { b.data = b.data[:0] }

// Compile-time check: fails to build if *Buffer stops satisfying it.
var _ Resetter = (*Buffer)(nil)

func Example_methodSets() {
	b := &Buffer{data: []string{"x"}}
	var r Resetter = b // fine: *Buffer has Reset
	// var r2 Resetter = Buffer{} would not compile:
	// Buffer does not implement Resetter (method Reset has pointer receiver)
	r.Reset()
	fmt.Println(len(b.data))

	// Output:
	// 0
}

// [end methodset]

// [start values]

func Example_interfaceValues() {
	var s Shape // nil interface: no type, no value
	fmt.Println(s == nil)

	s = Rect{1, 2}
	fmt.Printf("%T %v\n", s, s) // dynamic type and value

	s = Circle{3}
	fmt.Printf("%T %v\n", s, s)

	// Output:
	// true
	// interfaces.Rect {1 2}
	// interfaces.Circle {3}
}

// [end values]

// [start compose]

type Reader interface{ Read() string }
type Writer interface{ Write(s string) }

// ReadWriter embeds two interfaces: it requires both method sets.
type ReadWriter interface {
	Reader
	Writer
}

type memFile struct{ sb strings.Builder }

func (m *memFile) Read() string   { return m.sb.String() }
func (m *memFile) Write(s string) { m.sb.WriteString(s) }

func Example_embeddedInterfaces() {
	var rw ReadWriter = &memFile{}
	rw.Write("hello ")
	rw.Write("gopher")
	fmt.Println(rw.Read())

	// Output:
	// hello gopher
}

// [end compose]
