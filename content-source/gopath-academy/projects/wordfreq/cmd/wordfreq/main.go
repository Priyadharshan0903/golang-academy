// Command wordfreq prints the most frequent words in text.
//
// Usage:
//
//	wordfreq [-n N] [file ...]
//
// With no files, wordfreq reads standard input.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Priyadharshan0903/gopath-academy/projects/wordfreq/internal/count"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wordfreq", flag.ContinueOnError)
	fs.SetOutput(stderr)
	n := fs.Int("n", 10, "number of words to show (0 for all)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	totals := make(map[string]int)
	add := func(r io.Reader) error {
		counts, err := count.Words(r)
		if err != nil {
			return err
		}
		for w, c := range counts {
			totals[w] += c
		}
		return nil
	}

	if fs.NArg() == 0 {
		if err := add(stdin); err != nil {
			fmt.Fprintln(stderr, "wordfreq:", err)
			return 1
		}
	}
	for _, name := range fs.Args() {
		if err := addFile(name, add); err != nil {
			fmt.Fprintln(stderr, "wordfreq:", err)
			return 1
		}
	}

	for _, e := range count.Top(totals, *n) {
		fmt.Fprintf(stdout, "%7d  %s\n", e.Count, e.Word)
	}
	return 0
}

// addFile opens one file and feeds it to add. It is a separate
// function so the deferred Close runs after each file, not at the end.
func addFile(name string, add func(io.Reader) error) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return add(f)
}
