package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
)

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

var errUsage = errors.New("usage")

func usage(w io.Writer) {
	fmt.Fprint(w, `usage: shortctl [global flags] <command> [flags] [args]

commands:
  create [-code CODE] URL   shorten URL, optionally with a custom code
  get CODE                  show the link for CODE
  delete CODE               delete the link for CODE
  version                   print the version of shortctl

global flags:
`)
}

// [start run]

func run(ctx context.Context, args []string, e env, stdout, stderr io.Writer) int {
	cfg, rest, err := loadConfig(args, e, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}
	if err != nil {
		fmt.Fprintln(stderr, "shortctl:", err)
		return exitUsage
	}
	if len(rest) == 0 {
		usage(stderr)
		return exitUsage
	}

	c := &client{base: cfg.Server, http: &http.Client{Timeout: cfg.Timeout}}
	out := &printer{w: stdout, json: cfg.Output == "json"}

	switch rest[0] {
	case "create":
		err = cmdCreate(ctx, c, out, rest[1:], stderr)
	case "get":
		err = cmdGet(ctx, c, out, rest[1:])
	case "delete":
		if len(rest) != 2 {
			err = fmt.Errorf("%w: delete needs exactly one CODE", errUsage)
		} else if err = c.delete(ctx, rest[1]); err == nil {
			err = out.print(map[string]string{"deleted": rest[1]}, "deleted "+rest[1])
		}
	case "version":
		err = out.print(map[string]string{"version": version()}, version())
	default:
		err = fmt.Errorf("%w: unknown command %q", errUsage, rest[0])
	}

	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return exitOK
	case errors.Is(err, errUsage):
		fmt.Fprintln(stderr, "shortctl:", err)
		return exitUsage
	default:
		fmt.Fprintln(stderr, "shortctl:", err)
		return exitError
	}
}

// [end run]

func cmdCreate(ctx context.Context, c *client, out *printer, args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(stderr)
	code := fs.String("code", "", "custom short code")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return fmt.Errorf("%w: %v", errUsage, err)
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("%w: create needs exactly one URL", errUsage)
	}
	l, err := c.create(ctx, fs.Arg(0), *code)
	if err != nil {
		return err
	}
	return out.print(l, l.ShortURL)
}

func cmdGet(ctx context.Context, c *client, out *printer, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: get needs exactly one CODE", errUsage)
	}
	l, err := c.get(ctx, args[0])
	if err != nil {
		return err
	}
	return out.print(l, fmt.Sprintf("%s -> %s (created %s)", l.ShortURL, l.URL, l.CreatedAt.Format("2006-01-02 15:04 MST")))
}

// printer writes either JSON (for scripts) or a line of text (for people).
type printer struct {
	w    io.Writer
	json bool
}

func (p *printer) print(v any, text string) error {
	if p.json {
		enc := json.NewEncoder(p.w)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	_, err := fmt.Fprintln(p.w, text)
	return err
}

// version reports the module version stamped by the Go toolchain:
// v1.2.3 after `go install ...@v1.2.3`, or a VCS-based version for
// local builds (lesson 14.4).
func version() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return "unknown"
}
