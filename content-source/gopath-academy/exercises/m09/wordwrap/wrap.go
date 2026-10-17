//go:build !bug_overflow && !bug_longword && !bug_empty && !bug_lastline

// Package wordwrap is a "test your tests" exercise: the code is given
// and correct. Your job is to write tests in wrap_test.go that catch
// every one of the buggy versions in the bug_*.go files.
package wordwrap

import "strings"

// Wrap splits text into lines of at most width characters (bytes),
// breaking only at whitespace. Runs of whitespace collapse to one space.
// A word longer than width is placed on its own line, unbroken.
// Empty or whitespace-only text returns nil. width must be at least 1.
func Wrap(text string, width int) []string {
	var lines []string
	var cur strings.Builder
	for _, w := range strings.Fields(text) {
		switch {
		case cur.Len() == 0:
			cur.WriteString(w)
		case cur.Len()+1+len(w) <= width:
			cur.WriteByte(' ')
			cur.WriteString(w)
		default:
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
		}
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}
