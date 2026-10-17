// Command jsonq reads JSON and answers simple questions about it. It is
// the example program for lesson 14.1: subcommands with the standard
// flag package, input from files or standard input, and exit codes.
//
//	jsonq keys [-sort] [FILE...]
//	jsonq get [-raw] PATH [FILE...]
//	jsonq pretty [-indent N] [FILE...]
//
// With no FILE, jsonq reads standard input.
package main

import (
	"context"
	"os"
	"os/signal"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
