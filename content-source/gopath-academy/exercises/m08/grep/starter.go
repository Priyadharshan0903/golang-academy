//go:build !solution

// Package grep practices io.Reader, io.Writer and bufio by building a
// tiny grep and a reader wrapper.
package grep

import (
	"io"
	"regexp"
)

// Options controls Grep's output.
type Options struct {
	LineNumbers bool // prefix each match with "N:" (1-based)
	Invert      bool // print lines that do NOT match
}

// Grep reads r line by line and writes every line matching re to w,
// followed by "\n". It returns the number of lines written.
//
// Use bufio.Scanner, and return the scanner's error if reading fails,
// or the error from writing if w fails.
func Grep(r io.Reader, w io.Writer, re *regexp.Regexp, opts Options) (int, error) {
	// TODO: implement me.
	return 0, nil
}

// CountingReader wraps an io.Reader and counts the bytes read through it.
type CountingReader struct {
	R io.Reader
	N int64 // total bytes read so far
}

// Read reads from R and adds the number of bytes read to N.
func (c *CountingReader) Read(p []byte) (int, error) {
	// TODO: implement me.
	return 0, io.EOF
}
