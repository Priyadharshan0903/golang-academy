package values

import (
	"errors"
	"fmt"
	"strconv"
)

// [start create]

func Example_creating() {
	err1 := errors.New("connection refused")
	err2 := fmt.Errorf("port %d is out of range", 70000)

	fmt.Println(err1)
	fmt.Println(err2.Error()) // error is an interface with one method: Error() string

	// Output:
	// connection refused
	// port 70000 is out of range
}

// [end create]

// [start handle]

func parsePort(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parse port %q: %w", s, err)
	}
	if n < 1 || n > 65535 {
		return 0, fmt.Errorf("port %d out of range 1-65535", n)
	}
	return n, nil
}

func Example_handling() {
	for _, in := range []string{"8080", "http", "70000"} {
		port, err := parsePort(in)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		fmt.Println("ok:", port)
	}

	// Output:
	// ok: 8080
	// error: parse port "http": strconv.Atoi: parsing "http": invalid syntax
	// error: port 70000 out of range 1-65535
}

// [end handle]

// [start custom]

// ValidationError carries structured data, not just a message.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Reason
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Reason: "must not be negative"}
	}
	return nil
}

func Example_customType() {
	err := validateAge(-1)
	fmt.Println(err)

	var ve *ValidationError
	if errors.As(err, &ve) { // recover the structured data
		fmt.Println("bad field:", ve.Field)
	}

	// Output:
	// age: must not be negative
	// bad field: age
}

// [end custom]
