//go:build bug_lastline

package wordwrap

import "strings"

// Wrap loses the last line when there is more than one line.
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
	if len(lines) == 0 && cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}
