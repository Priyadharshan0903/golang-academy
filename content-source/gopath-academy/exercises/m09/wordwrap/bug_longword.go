//go:build bug_longword

package wordwrap

import "strings"

// Wrap silently drops words longer than width.
func Wrap(text string, width int) []string {
	var lines []string
	var cur strings.Builder
	for _, w := range strings.Fields(text) {
		if len(w) > width {
			continue
		}
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
