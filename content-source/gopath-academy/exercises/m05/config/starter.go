//go:build !solution

// Package config parses simple "key = value" configuration files and
// practices wrapping errors with context that callers can inspect.
package config

import (
	"errors"
	"io"
)

// ErrSyntax is wrapped by every error caused by a malformed line.
var ErrSyntax = errors.New("syntax error")

// ErrDuplicateKey is wrapped when a key appears twice.
var ErrDuplicateKey = errors.New("duplicate key")

// ParseError says which line of the input failed, and why.
type ParseError struct {
	Line int   // 1-based line number
	Err  error // the underlying cause
}

// Error formats the error as "line N: cause".
func (e *ParseError) Error() string {
	// TODO: implement me.
	return ""
}

// Unwrap returns the underlying cause, so errors.Is and errors.As
// can look through a ParseError.
func (e *ParseError) Unwrap() error {
	// TODO: implement me.
	return nil
}

// ParseLine parses one "key = value" line. Spaces around the key and
// value are trimmed. A line without "=" or with an empty key returns
// an error that wraps ErrSyntax and mentions the line's text.
func ParseLine(line string) (key, value string, err error) {
	// TODO: implement me.
	return "", "", nil
}

// Load reads every line of r. Blank lines and lines starting with "#"
// (after trimming spaces) are ignored. The first bad line stops
// parsing and is returned as a *ParseError whose Err wraps either
// ErrSyntax or ErrDuplicateKey.
func Load(r io.Reader) (map[string]string, error) {
	// TODO: implement me.
	return nil, nil
}
