//go:build !solution

// Package cli parses command line arguments for a tiny greeter.
//
// Keeping parsing in a function that takes []string (instead of reading
// os.Args directly) makes it easy to test. main would call
// cli.Parse(os.Args[1:]).
package cli

import (
	"errors"
	"flag"
	"io"
)

// Config holds the parsed options.
type Config struct {
	Name  string // -name, default "Gopher"
	Shout bool   // -shout, default false
	Times int    // -times, default 1, must be >= 1
}

// ErrInvalidTimes is returned when -times is less than 1.
var ErrInvalidTimes = errors.New("times must be at least 1")

// Parse parses args (without the program name) into a Config.
//
// Requirements:
//   - Use flag.NewFlagSet("greet", flag.ContinueOnError), not the global
//     flag package functions, so tests do not share state.
//   - Call fs.SetOutput(io.Discard) so errors are returned, not printed.
//   - Return ErrInvalidTimes if -times < 1.
//   - Return the flag package's error for unknown flags.
func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("greet", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	// TODO: define the three flags, parse args, validate, return.
	_ = fs
	return Config{}, nil
}

// Render returns the greeting text for cfg: "Hello, <Name>!" repeated
// Times times, one per line, with no trailing newline. If Shout is true
// the whole text is uppercase.
//
// Hint: strings.Repeat, strings.Join and strings.ToUpper are useful.
func Render(cfg Config) string {
	// TODO: implement me.
	return ""
}
