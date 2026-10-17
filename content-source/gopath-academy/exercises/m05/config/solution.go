//go:build solution

// Package config parses simple "key = value" configuration files and
// practices wrapping errors with context that callers can inspect.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrSyntax is wrapped by every error caused by a malformed line.
var ErrSyntax = errors.New("syntax error")

// ErrDuplicateKey is wrapped when a key appears twice.
var ErrDuplicateKey = errors.New("duplicate key")

// ParseError says which line of the input failed, and why.
type ParseError struct {
	Line int
	Err  error
}

// Error formats the error as "line N: cause".
func (e *ParseError) Error() string { return fmt.Sprintf("line %d: %v", e.Line, e.Err) }

// Unwrap returns the underlying cause.
func (e *ParseError) Unwrap() error { return e.Err }

// ParseLine parses one "key = value" line.
func ParseLine(line string) (key, value string, err error) {
	k, v, ok := strings.Cut(line, "=")
	k = strings.TrimSpace(k)
	if !ok || k == "" {
		return "", "", fmt.Errorf("%w: expected key = value, got %q", ErrSyntax, line)
	}
	return k, strings.TrimSpace(v), nil
}

// Load reads every line of r into a map.
func Load(r io.Reader) (map[string]string, error) {
	cfg := make(map[string]string)
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, err := ParseLine(line)
		if err != nil {
			return nil, &ParseError{Line: n, Err: err}
		}
		if _, dup := cfg[k]; dup {
			return nil, &ParseError{Line: n, Err: fmt.Errorf("%w: %q", ErrDuplicateKey, k)}
		}
		cfg[k] = v
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return cfg, nil
}
