//go:build !solution

package todo

import "io"

// Run is the whole program; main would call
// os.Exit(Run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr)).
//
// Usage:
//
//	todo [-file PATH] add TEXT...     add an item (the words are joined with spaces)
//	todo [-file PATH] list [-all]     print open items as "N. text"; -all also
//	                                  prints done items, as "N. [x] text"
//	todo [-file PATH] done N          mark item N (1-based, as printed by list) done
//
// The list lives in the -file flag's path; if the flag is not given, in
// $TODO_FILE; if that is empty too, in "todo.json".
//
// Exit codes: 0 on success, 1 for runtime errors (such as an unreadable
// file or an item number that does not exist), 2 for usage errors
// (unknown command, bad flags, missing or non-numeric arguments).
// Errors are printed to stderr, prefixed with "todo: ".
func Run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	// TODO: implement me. load and save are in store.go.
	return 2
}
