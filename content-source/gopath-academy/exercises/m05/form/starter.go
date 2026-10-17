//go:build !solution

// Package form validates user input and reports every problem at once
// with errors.Join, instead of stopping at the first one.
package form

import "errors"

// Sentinel errors for each kind of problem.
var (
	ErrRequired = errors.New("is required")
	ErrTooShort = errors.New("is too short")
	ErrInvalid  = errors.New("is invalid")
)

// FieldError reports a problem with one field.
type FieldError struct {
	Field string
	Err   error // one of the sentinel errors above
}

// Error formats the error as "<field> <reason>", for example
// "password is too short".
func (e *FieldError) Error() string {
	// TODO: implement me.
	return ""
}

// Unwrap returns the sentinel error.
func (e *FieldError) Unwrap() error {
	// TODO: implement me.
	return nil
}

// Signup is the input of a signup form.
type Signup struct {
	Username string
	Email    string
	Password string
}

// Validate checks every field and returns nil when the form is valid.
// Otherwise it returns errors.Join of one *FieldError per problem, in
// this order:
//
//   - username: ErrRequired if empty
//   - email:    ErrRequired if empty, otherwise ErrInvalid if it has no "@"
//   - password: ErrRequired if empty, otherwise ErrTooShort if shorter than 8 bytes
func Validate(s Signup) error {
	// TODO: implement me.
	return nil
}

// Fields returns the names of the fields that failed, in order, or
// nil when err is nil. Use errors.As on each joined error.
//
// Hint: an error made by errors.Join has a method
// Unwrap() []error. Assert to an interface with that method.
func Fields(err error) []string {
	// TODO: implement me.
	return nil
}
