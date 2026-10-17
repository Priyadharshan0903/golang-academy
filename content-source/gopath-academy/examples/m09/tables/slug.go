package tables

import (
	"strings"
	"unicode"
)

// [start code]

// Slugify turns a title into a URL-friendly slug:
// lowercase letters and digits separated by single hyphens.
func Slugify(title string) string {
	var b strings.Builder
	pendingHyphen := false
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if pendingHyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			pendingHyphen = false
			b.WriteRune(r)
		} else {
			pendingHyphen = true
		}
	}
	return b.String()
}

// [end code]
