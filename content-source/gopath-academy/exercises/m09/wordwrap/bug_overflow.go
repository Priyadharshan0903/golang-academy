//go:build bug_overflow

package wordwrap

import "strings"

// Wrap has an off-by-one: it allows lines one character longer than width.
func Wrap(text string, width int) []string {
	var lines []string
	var cur strings.Builder
	for _, w := range strings.Fields(text) {
		switch {
		case cur.Len() == 0:
			cur.WriteString(w)
		case cur.Len()+len(w) <= width: // forgot the space
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
