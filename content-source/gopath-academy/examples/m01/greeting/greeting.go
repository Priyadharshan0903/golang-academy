// [start doc]

// Package greeting builds friendly greetings.
//
// It exists to show how Go decides what is visible outside a package:
// names that start with an uppercase letter are exported, everything
// else is private to the package.
package greeting

// [end doc]

// [start visibility]

import "strings"

// DefaultName is used when the caller passes an empty name.
// It is exported because it starts with an uppercase letter.
const DefaultName = "Gopher"

// Hello returns a greeting for name. Leading and trailing spaces are
// ignored, and an empty name falls back to DefaultName.
func Hello(name string) string {
	return "Hello, " + normalize(name) + "!"
}

// normalize is unexported: code outside this package cannot call it.
func normalize(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return DefaultName
	}
	return name
}

// [end visibility]
