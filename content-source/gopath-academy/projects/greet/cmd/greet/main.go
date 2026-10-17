// Command greet prints a friendly greeting.
//
// Usage:
//
//	greet [-name NAME] [-shout] [-times N] [-version]
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Priyadharshan0903/gopath-academy/projects/greet/internal/greeter"
	"github.com/Priyadharshan0903/gopath-academy/projects/greet/internal/version"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run holds the whole program. main only connects it to the real
// process (arguments, output streams, exit code), which makes run
// easy to test.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("greet", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var opts greeter.Options
	fs.StringVar(&opts.Name, "name", "Gopher", "who to greet")
	fs.BoolVar(&opts.Shout, "shout", false, "print in uppercase")
	fs.IntVar(&opts.Times, "times", 1, "how many times to greet")
	showVersion := fs.Bool("version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		return 2 // the flag package already printed the problem
	}
	if *showVersion {
		fmt.Fprintln(stdout, version.String())
		return 0
	}
	if opts.Times < 1 {
		fmt.Fprintln(stderr, "greet: -times must be at least 1")
		return 2
	}
	fmt.Fprintln(stdout, greeter.Render(opts))
	return 0
}
