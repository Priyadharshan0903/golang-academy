//go:build bug_empty

package wordwrap

import "strings"

// Wrap returns [""] instead of nil for empty input.
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
	return append(lines, cur.String())
}
