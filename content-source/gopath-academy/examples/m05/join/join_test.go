package join

import (
	"errors"
	"fmt"
	"strings"
)

// [start join]

var (
	ErrNameRequired = errors.New("name is required")
	ErrBadEmail     = errors.New("email is invalid")
)

type Signup struct{ Name, Email string }

// Validate reports every problem at once instead of stopping at the first.
func (s Signup) Validate() error {
	var errs []error
	if s.Name == "" {
		errs = append(errs, ErrNameRequired)
	}
	if !strings.Contains(s.Email, "@") {
		errs = append(errs, fmt.Errorf("%w: %q", ErrBadEmail, s.Email))
	}
	return errors.Join(errs...) // nil if errs is empty
}

func Example_join() {
	err := Signup{Email: "nope"}.Validate()
	fmt.Println(err) // one error per line
	fmt.Println(errors.Is(err, ErrNameRequired), errors.Is(err, ErrBadEmail))

	fmt.Println(Signup{Name: "Ana", Email: "a@b.c"}.Validate() == nil)

	// Output:
	// name is required
	// email is invalid: "nope"
	// true true
	// true
}

// [end join]

// [start multiw]

func Example_multipleW() {
	errA := errors.New("cache miss")
	errB := errors.New("db timeout")

	// fmt.Errorf accepts several %w verbs (Go 1.20+).
	err := fmt.Errorf("load user: %w, then %w", errA, errB)
	fmt.Println(err)
	fmt.Println(errors.Is(err, errA), errors.Is(err, errB))

	// errors.Unwrap only follows single-error wrappers, so it returns nil here.
	fmt.Println(errors.Unwrap(err) == nil)

	// Output:
	// load user: cache miss, then db timeout
	// true true
	// true
}

// [end multiw]
