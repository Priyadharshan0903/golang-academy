//go:build solution

// Package grep practices io.Reader, io.Writer and bufio by building a
// tiny grep and a reader wrapper.
package grep

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
)

// Options controls Grep's output.
type Options struct {
	LineNumbers bool
	Invert      bool
}

// Grep writes the lines of r that match re to w.
func Grep(r io.Reader, w io.Writer, re *regexp.Regexp, opts Options) (int, error) {
	bw := bufio.NewWriter(w)
	sc := bufio.NewScanner(r)
	written := 0
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		if re.MatchString(line) == opts.Invert {
			continue
		}
		var err error
		if opts.LineNumbers {
			_, err = fmt.Fprintf(bw, "%d:%s\n", n, line)
		} else {
			_, err = fmt.Fprintln(bw, line)
		}
		if err != nil {
			return written, err
		}
		written++
	}
	if err := sc.Err(); err != nil {
		return written, err
	}
	return written, bw.Flush()
}

// CountingReader wraps an io.Reader and counts the bytes read through it.
type CountingReader struct {
	R io.Reader
	N int64
}

// Read reads from R and adds the number of bytes read to N.
func (c *CountingReader) Read(p []byte) (int, error) {
	n, err := c.R.Read(p)
	c.N += int64(n)
	return n, err
}
