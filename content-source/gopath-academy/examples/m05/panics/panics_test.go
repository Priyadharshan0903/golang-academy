package panics

import (
	"errors"
	"fmt"
	"regexp"
)

// [start runtime]

func Example_runtimePanics() {
	try := func(name string, f func()) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("%s: %v\n", name, r)
			}
		}()
		f()
	}

	try("index", func() {
		s := []int{1, 2, 3}
		i := 5
		_ = s[i]
	})
	try("nil map", func() {
		var m map[string]int
		m["x"] = 1
	})
	try("explicit", func() { panic("something impossible happened") })

	// Output:
	// index: runtime error: index out of range [5] with length 3
	// nil map: assignment to entry in nil map
	// explicit: something impossible happened
}

// [end runtime]

// [start must]

// MustCompile-style helpers panic on bad input. They are meant for
// values fixed at program start, where an error is a programming bug.
var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func Example_mustHelpers() {
	fmt.Println(datePattern.MatchString("2026-10-16"))

	// For user-supplied patterns, use the error-returning form.
	_, err := regexp.Compile(`[unclosed`)
	fmt.Println(err != nil)

	// Output:
	// true
	// true
}

// [end must]

// [start boundary]

// safeCall converts a panic inside f into an ordinary error. Do this
// only at boundaries (a request handler, a plugin call, a worker),
// never as a substitute for normal error handling.
func safeCall(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	f()
	return nil
}

func Example_recoverAtBoundary() {
	err := safeCall(func() {
		var p *struct{ N int }
		fmt.Println(p.N) // nil dereference
	})
	fmt.Println(err)
	fmt.Println(safeCall(func() {}) == nil)

	// Output:
	// panic: runtime error: invalid memory address or nil pointer dereference
	// true
}

// [end boundary]

// [start repanic]

var errFatal = errors.New("fatal config error")

// handle recovers only the panics it understands and re-panics others.
func handle(f func()) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if e, ok := r.(error); ok && errors.Is(e, errFatal) {
			err = e
			return
		}
		panic(r) // not ours: let it crash loudly
	}()
	f()
	return nil
}

func Example_selectiveRecover() {
	fmt.Println(handle(func() { panic(errFatal) }))

	defer func() { fmt.Println("outer caught:", recover()) }()
	handle(func() { panic("unexpected") })

	// Output:
	// fatal config error
	// outer caught: unexpected
}

// [end repanic]
