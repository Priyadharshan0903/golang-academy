package nilinterface

import (
	"errors"
	"fmt"
)

// [start trap]

type MyError struct{ Msg string }

func (e *MyError) Error() string { return e.Msg }

// validate returns a nil *MyError when everything is fine...
func validateBroken(ok bool) error {
	var err *MyError // nil pointer of type *MyError
	if !ok {
		err = &MyError{Msg: "invalid"}
	}
	return err // ...wrapped in a NON-nil error interface!
}

func validate(ok bool) error {
	if !ok {
		return &MyError{Msg: "invalid"}
	}
	return nil // a real nil interface
}

func Example_nilInterfaceTrap() {
	err := validateBroken(true)
	fmt.Println(err == nil) // false: the interface holds (type=*MyError, value=nil)
	fmt.Printf("%T %v\n", err, err == (*MyError)(nil))

	fmt.Println(validate(true) == nil)

	// Output:
	// false
	// *nilinterface.MyError true
	// true
}

// [end trap]

// [start pair]

func Example_typeValuePair() {
	var e1 error                 // (nil type, nil value)  -> == nil
	var p *MyError               // a nil pointer
	var e2 error = p             // (*MyError, nil)        -> != nil
	var e3 error = &MyError{"x"} // (*MyError, pointer)   -> != nil

	fmt.Println(e1 == nil, e2 == nil, e3 == nil)

	// errors.As still finds the typed nil inside e2.
	var target *MyError
	fmt.Println(errors.As(e2, &target), target == nil)

	// Output:
	// true false false
	// true true
}

// [end pair]
