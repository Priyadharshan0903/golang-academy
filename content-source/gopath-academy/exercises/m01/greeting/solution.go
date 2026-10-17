//go:build solution

// Package greeting is your first exercise.
package greeting

import "strings"

// Greeting returns "Hello, <name>!", falling back to "Gopher".
func Greeting(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Gopher"
	}
	return "Hello, " + name + "!"
}
