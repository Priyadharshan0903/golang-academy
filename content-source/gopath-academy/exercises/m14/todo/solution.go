//go:build solution

package todo

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var errUsage = errors.New("usage")

// Run is the whole program. See the starter for the specification.
func Run(args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	err := run(args, getenv, stdout, stderr)
	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return 0
	case errors.Is(err, errUsage):
		fmt.Fprintln(stderr, "todo:", err)
		return 2
	default:
		fmt.Fprintln(stderr, "todo:", err)
		return 1
	}
}

func run(args []string, getenv func(string) string, stdout, stderr io.Writer) error {
	global := flag.NewFlagSet("todo", flag.ContinueOnError)
	global.SetOutput(stderr)
	file := global.String("file", "", "path of the list (default $TODO_FILE or todo.json)")
	if err := global.Parse(args); err != nil {
		return usage(err)
	}
	path := *file
	if path == "" {
		path = getenv("TODO_FILE")
	}
	if path == "" {
		path = "todo.json"
	}

	rest := global.Args()
	if len(rest) == 0 {
		return fmt.Errorf("%w: expected a command: add, list or done", errUsage)
	}
	cmd, cmdArgs := rest[0], rest[1:]

	items, err := load(path)
	if err != nil {
		return err
	}

	switch cmd {
	case "add":
		if len(cmdArgs) == 0 {
			return fmt.Errorf("%w: add needs some text", errUsage)
		}
		items = append(items, Item{Text: strings.Join(cmdArgs, " ")})
		if err := save(path, items); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "added %d. %s\n", len(items), items[len(items)-1].Text)
		return nil

	case "list":
		fs := flag.NewFlagSet("list", flag.ContinueOnError)
		fs.SetOutput(stderr)
		all := fs.Bool("all", false, "include done items")
		if err := fs.Parse(cmdArgs); err != nil {
			return usage(err)
		}
		for i, it := range items {
			switch {
			case !it.Done:
				fmt.Fprintf(stdout, "%d. %s\n", i+1, it.Text)
			case *all:
				fmt.Fprintf(stdout, "%d. [x] %s\n", i+1, it.Text)
			}
		}
		return nil

	case "done":
		if len(cmdArgs) != 1 {
			return fmt.Errorf("%w: done needs exactly one item number", errUsage)
		}
		n, err := strconv.Atoi(cmdArgs[0])
		if err != nil {
			return fmt.Errorf("%w: %q is not a number", errUsage, cmdArgs[0])
		}
		if n < 1 || n > len(items) {
			return fmt.Errorf("no item %d (the list has %d)", n, len(items))
		}
		items[n-1].Done = true
		return save(path, items)

	default:
		return fmt.Errorf("%w: unknown command %q", errUsage, cmd)
	}
}

func usage(err error) error {
	if errors.Is(err, flag.ErrHelp) {
		return err
	}
	return fmt.Errorf("%w: %v", errUsage, err)
}
