//go:build solution

// Package cli parses command line arguments for a tiny greeter.
package cli

import (
	"errors"
	"flag"
	"io"
	"strings"
)

// Config holds the parsed options.
type Config struct {
	Name  string
	Shout bool
	Times int
}

// ErrInvalidTimes is returned when -times is less than 1.
var ErrInvalidTimes = errors.New("times must be at least 1")

// Parse parses args (without the program name) into a Config.
func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("greet", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var cfg Config
	fs.StringVar(&cfg.Name, "name", "Gopher", "who to greet")
	fs.BoolVar(&cfg.Shout, "shout", false, "print in uppercase")
	fs.IntVar(&cfg.Times, "times", 1, "how many times to greet")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if cfg.Times < 1 {
		return Config{}, ErrInvalidTimes
	}
	return cfg, nil
}

// Render returns the greeting text for cfg.
func Render(cfg Config) string {
	lines := make([]string, cfg.Times)
	for i := range lines {
		lines[i] = "Hello, " + cfg.Name + "!"
	}
	out := strings.Join(lines, "\n")
	if cfg.Shout {
		out = strings.ToUpper(out)
	}
	return out
}
