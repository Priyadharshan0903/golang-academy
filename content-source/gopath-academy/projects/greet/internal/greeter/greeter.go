// Package greeter contains the logic of the greet command.
//
// It lives under internal/, so only code inside projects/greet can
// import it. That keeps it free to change without breaking anyone.
package greeter

import "strings"

// Options controls how a greeting is rendered.
type Options struct {
	Name  string
	Shout bool
	Times int
}

// Render returns the greeting described by opts. An empty Name becomes
// "Gopher" and a Times value below 1 is treated as 1.
func Render(opts Options) string {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = "Gopher"
	}
	times := max(opts.Times, 1)

	lines := make([]string, times)
	for i := range lines {
		lines[i] = "Hello, " + name + "!"
	}
	out := strings.Join(lines, "\n")
	if opts.Shout {
		out = strings.ToUpper(out)
	}
	return out
}
