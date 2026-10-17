package assertions

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// [start assert]

func Example_typeAssertion() {
	var v any = "hello"

	s := v.(string) // panics if v does not hold a string
	fmt.Println(s)

	n, ok := v.(int) // comma-ok form never panics
	fmt.Println(n, ok)

	// Output:
	// hello
	// 0 false
}

// [end assert]

// [start switch]

func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int, int64:
		return fmt.Sprintf("an integer: %v", x) // x is any here: two types listed
	case string:
		return fmt.Sprintf("a string of length %d", len(x)) // x is a string
	case error:
		return "an error: " + x.Error()
	case fmt.Stringer:
		return "a Stringer: " + x.String()
	default:
		return fmt.Sprintf("something else: %T", x)
	}
}

type point struct{ X, Y int }

func (p point) String() string { return fmt.Sprintf("(%d,%d)", p.X, p.Y) }

func Example_typeSwitch() {
	fmt.Println(describe(nil))
	fmt.Println(describe(42))
	fmt.Println(describe("gopher"))
	fmt.Println(describe(errors.New("boom")))
	fmt.Println(describe(point{1, 2}))
	fmt.Println(describe(3.14))

	// Output:
	// nil
	// an integer: 42
	// a string of length 6
	// an error: boom
	// a Stringer: (1,2)
	// something else: float64
}

// [end switch]

// [start capability]

// Asserting to an interface checks for an optional capability.
// io.Copy does exactly this to use faster paths when it can.
func describeReader(r io.Reader) string {
	if _, ok := r.(io.Seeker); ok {
		return "seekable reader"
	}
	return "stream-only reader"
}

func Example_capabilityCheck() {
	f, _ := os.Open(os.DevNull)
	defer f.Close()

	fmt.Println(describeReader(strings.NewReader("x")))
	fmt.Println(describeReader(f))
	fmt.Println(describeReader(io.LimitReader(f, 10)))

	// Output:
	// seekable reader
	// seekable reader
	// stream-only reader
}

// [end capability]
