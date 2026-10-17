package deferring

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// [start order]

func Example_order() {
	fmt.Println("start")
	defer fmt.Println("deferred 1")
	defer fmt.Println("deferred 2") // runs first: last in, first out
	fmt.Println("end")

	// Output:
	// start
	// end
	// deferred 2
	// deferred 1
}

// [end order]

// [start args]

func Example_argumentsEvaluatedEarly() {
	x := 1
	defer fmt.Println("deferred sees x =", x) // x is evaluated NOW
	x = 2
	fmt.Println("x is now", x)

	// Output:
	// x is now 2
	// deferred sees x = 1
}

// [end args]

// [start closure]

func Example_closureSeesLatest() {
	x := 1
	defer func() { fmt.Println("closure sees x =", x) }() // reads x later
	x = 2

	// Output:
	// closure sees x = 2
}

// [end closure]

// [start cleanup]

// firstLine shows the standard pattern: acquire, check the error,
// then defer the release right away.
func firstLine(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close() // runs on every return path below

	buf := make([]byte, 64)
	n, err := f.Read(buf)
	if err != nil {
		return "", err
	}
	for i, b := range buf[:n] {
		if b == '\n' {
			return string(buf[:i]), nil
		}
	}
	return string(buf[:n]), nil
}

func Example_cleanup() {
	dir, _ := os.MkdirTemp("", "defer-example")
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "notes.txt")
	_ = os.WriteFile(path, []byte("hello\nworld\n"), 0o644)

	line, err := firstLine(path)
	fmt.Println(line, err)

	_, err = firstLine(filepath.Join(dir, "missing.txt"))
	fmt.Println(errors.Is(err, os.ErrNotExist))

	// Output:
	// hello <nil>
	// true
}

// [end cleanup]

// [start named]

// safeDivide turns a panic into an error by changing the named
// result from a deferred function. Panics are covered in Module 5.
func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return a / b, nil
}

func Example_modifyNamedResult() {
	fmt.Println(safeDivide(10, 2))
	fmt.Println(safeDivide(1, 0))

	// Output:
	// 5 <nil>
	// 0 recovered: runtime error: integer divide by zero
}

// [end named]
