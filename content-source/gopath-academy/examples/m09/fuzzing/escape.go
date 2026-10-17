package fuzzing

import (
	"errors"
	"strings"
)

// [start code]

// Quote wraps s in double quotes, escaping backslashes and quotes.
func Quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"', '\\':
			b.WriteByte('\\')
		}
		b.WriteByte(s[i])
	}
	b.WriteByte('"')
	return b.String()
}

var ErrSyntax = errors.New("invalid quoted string")

// Unquote reverses Quote.
func Unquote(q string) (string, error) {
	if len(q) < 2 || q[0] != '"' || q[len(q)-1] != '"' {
		return "", ErrSyntax
	}
	var b strings.Builder
	body := q[1 : len(q)-1]
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c == '\\' {
			i++
			if i == len(body) {
				return "", ErrSyntax // dangling backslash
			}
			c = body[i]
		} else if c == '"' {
			return "", ErrSyntax // unescaped quote inside
		}
		b.WriteByte(c)
	}
	return b.String(), nil
}

// [end code]
