package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Exit codes. 2 for usage errors is the convention of the flag package
// and of most Unix tools.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// [start run]

// command is one subcommand: its flags and what it does.
type command struct {
	name    string
	summary string
	run     func(ctx context.Context, args []string, env *env) error
}

// env is what commands may touch. Tests provide buffers instead of the
// real standard streams.
type env struct {
	stdin          io.Reader
	stdout, stderr io.Writer
}

var commands = []command{
	{"keys", "print the top-level keys of each object", runKeys},
	{"get", "print the value at a dotted PATH, such as user.emails.0", runGet},
	{"pretty", "re-indent JSON", runPretty},
}

// errUsage marks errors that are the caller's fault (exit code 2).
var errUsage = errors.New("usage error")

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	e := &env{stdin: stdin, stdout: stdout, stderr: stderr}
	if len(args) == 0 || args[0] == "-h" || args[0] == "-help" || args[0] == "help" {
		usage(stderr)
		return exitUsage
	}
	i := slices.IndexFunc(commands, func(c command) bool { return c.name == args[0] })
	if i < 0 {
		fmt.Fprintf(stderr, "jsonq: unknown command %q\n\n", args[0])
		usage(stderr)
		return exitUsage
	}

	err := commands[i].run(ctx, args[1:], e)
	switch {
	case err == nil:
		return exitOK
	case errors.Is(err, flag.ErrHelp):
		return exitOK // -h was requested and printed
	case errors.Is(err, errUsage):
		fmt.Fprintln(stderr, "jsonq:", err)
		return exitUsage
	default:
		fmt.Fprintln(stderr, "jsonq:", err)
		return exitError
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: jsonq <command> [flags] [args]\n\ncommands:")
	for _, c := range commands {
		fmt.Fprintf(w, "  %-8s %s\n", c.name, c.summary)
	}
	fmt.Fprintln(w, "\nRun 'jsonq <command> -h' for the flags of a command.")
}

// [end run]

// [start flagset]

// newFlags returns a FlagSet that reports errors instead of exiting,
// and prints its usage to the command's stderr.
func newFlags(name string, e *env, argsHelp string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(e.stderr)
	fs.Usage = func() {
		fmt.Fprintf(e.stderr, "usage: jsonq %s [flags] %s\n", name, argsHelp)
		fs.PrintDefaults()
	}
	return fs
}

func parse(fs *flag.FlagSet, args []string) error {
	err := fs.Parse(args)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	return err
}

// [end flagset]

// [start commands]

func runKeys(ctx context.Context, args []string, e *env) error {
	fs := newFlags("keys", e, "[FILE...]")
	sorted := fs.Bool("sort", false, "sort the keys")
	if err := parse(fs, args); err != nil {
		return err
	}
	return eachDocument(ctx, fs.Args(), e, func(name string, v any) error {
		obj, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: not a JSON object", name)
		}
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		if *sorted {
			slices.Sort(keys)
		}
		fmt.Fprintln(e.stdout, strings.Join(keys, "\n"))
		return nil
	})
}

func runGet(ctx context.Context, args []string, e *env) error {
	fs := newFlags("get", e, "PATH [FILE...]")
	raw := fs.Bool("raw", false, "print strings without quotes")
	if err := parse(fs, args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("%w: missing PATH", errUsage)
	}
	path := fs.Arg(0)
	return eachDocument(ctx, fs.Args()[1:], e, func(name string, v any) error {
		got, err := lookup(v, path)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if s, ok := got.(string); ok && *raw {
			fmt.Fprintln(e.stdout, s)
			return nil
		}
		b, _ := json.Marshal(got)
		fmt.Fprintln(e.stdout, string(b))
		return nil
	})
}

func runPretty(ctx context.Context, args []string, e *env) error {
	fs := newFlags("pretty", e, "[FILE...]")
	indent := fs.Int("indent", 2, "spaces per level (0 for compact)")
	if err := parse(fs, args); err != nil {
		return err
	}
	if *indent < 0 || *indent > 8 {
		return fmt.Errorf("%w: -indent must be between 0 and 8", errUsage)
	}
	return eachDocument(ctx, fs.Args(), e, func(name string, v any) error {
		enc := json.NewEncoder(e.stdout)
		enc.SetIndent("", strings.Repeat(" ", *indent))
		return enc.Encode(v)
	})
}

// [end commands]

// [start input]

// eachDocument calls fn for every JSON value in the named files, or in
// standard input if there are none. A file may hold several values,
// one after another, like JSON Lines.
func eachDocument(ctx context.Context, files []string, e *env, fn func(name string, v any) error) error {
	if len(files) == 0 {
		return decodeAll(ctx, "stdin", e.stdin, fn)
	}
	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			return err // already says which file
		}
		err = decodeAll(ctx, name, f, fn)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeAll(ctx context.Context, name string, r io.Reader, fn func(string, any) error) error {
	dec := json.NewDecoder(r)
	dec.UseNumber() // keep numbers exactly as written
	for {
		if err := ctx.Err(); err != nil {
			return err // Ctrl+C stops between documents
		}
		var v any
		err := dec.Decode(&v)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: invalid JSON: %w", name, err)
		}
		if err := fn(name, v); err != nil {
			return err
		}
	}
}

// [end input]

// lookup follows a dotted path through objects and arrays.
func lookup(v any, path string) (any, error) {
	for _, part := range strings.Split(path, ".") {
		switch node := v.(type) {
		case map[string]any:
			next, ok := node[part]
			if !ok {
				return nil, fmt.Errorf("no key %q", part)
			}
			v = next
		case []any:
			i, err := strconv.Atoi(part)
			if err != nil || i < 0 || i >= len(node) {
				return nil, fmt.Errorf("no index %q in an array of %d", part, len(node))
			}
			v = node[i]
		default:
			return nil, fmt.Errorf("cannot look up %q in a %T", part, v)
		}
	}
	return v, nil
}
