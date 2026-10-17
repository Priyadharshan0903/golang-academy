//go:build solution

// Package form validates user input and reports every problem at once
// with errors.Join, instead of stopping at the first one.
package form

import (
	"errors"
	"strings"
)

// Sentinel errors for each kind of problem.
var (
	ErrRequired = errors.New("is required")
	ErrTooShort = errors.New("is too short")
	ErrInvalid  = errors.New("is invalid")
)

// FieldError reports a problem with one field.
type FieldError struct {
	Field string
	Err   error
}

// Error formats the error as "<field> <reason>".
func (e *FieldError) Error() string { return e.Field + " " + e.Err.Error() }

// Unwrap returns the sentinel error.
func (e *FieldError) Unwrap() error { return e.Err }

// Signup is the input of a signup form.
type Signup struct {
	Username string
	Email    string
	Password string
}

// Validate checks every field and joins all problems.
func Validate(s Signup) error {
	var errs []error
	add := func(field string, err error) {
		errs = append(errs, &FieldError{Field: field, Err: err})
	}

	if s.Username == "" {
		add("username", ErrRequired)
	}
	switch {
	case s.Email == "":
		add("email", ErrRequired)
	case !strings.Contains(s.Email, "@"):
		add("email", ErrInvalid)
	}
	switch {
	case s.Password == "":
		add("password", ErrRequired)
	case len(s.Password) < 8:
		add("password", ErrTooShort)
	}
	return errors.Join(errs...)
}

// Fields returns the names of the fields that failed, in order.
func Fields(err error) []string {
	if err == nil {
		return nil
	}
	var list []error
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		list = joined.Unwrap()
	} else {
		list = []error{err}
	}
	var names []string
	for _, e := range list {
		var fe *FieldError
		if errors.As(e, &fe) {
			names = append(names, fe.Field)
		}
	}
	return names
}
