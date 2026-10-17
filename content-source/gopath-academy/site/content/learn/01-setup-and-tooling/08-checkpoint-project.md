---
title: "1.8 Checkpoint project: the greet CLI"
linkTitle: "1.8 Checkpoint: greet"
weight: 8
description: "Build a small, properly structured command line tool with flags, tests, version stamping and cross-compiled releases."
goVersion: "1.27"
---

## What you will build

A command line tool called `greet`:

```bash
greet -name Ana -times 2 -shout
```

```text
HELLO, ANA!
HELLO, ANA!
```

It is deliberately small, so you can focus on structure and tooling instead of logic. By the end you will have used everything from this module: a module, several packages, `internal/`, `cmd/`, flags, tests, `-ldflags` version stamping and cross-compiling.

The finished code is in [`projects/greet`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/greet). Try building it yourself first, using the steps below, and compare afterwards.

## The plan

{{% steps %}}

### Lay out the project

Use the layout from lesson 1.5: one binary in `cmd/greet`, the logic in `internal/greeter`, build metadata in `internal/version`.

### Write the logic as a plain function

The greeting logic has no idea it is part of a CLI. It takes options and returns a string, which makes it trivial to test.

{{< example file="projects/greet/internal/greeter/greeter.go" >}}

`max` is a built-in function since Go 1.21, so `max(opts.Times, 1)` clamps the value without an `if`.

### Test it with a table

{{< example file="projects/greet/internal/greeter/greeter_test.go" run="go test ./projects/greet/..." >}}

This is a **table-driven test**: a slice of cases and one loop. It is the most common testing style in Go, and Module 9 covers it in depth.

### Keep `main` thin, and testable

`main` should only connect the program to the real world. Everything else goes in a `run` function that takes its arguments and output streams as parameters and returns an exit code.

{{< example file="projects/greet/cmd/greet/main.go" >}}

A few details worth noticing:

- `flag.NewFlagSet(..., flag.ContinueOnError)` returns errors instead of calling `os.Exit` itself, so `run` stays in control.
- `fs.SetOutput(stderr)` sends usage and error messages to the stream we pass in, not straight to the terminal.
- Exit code `2` for bad usage and `0` for success follow Unix conventions (and match what the `flag` package does by default).

### Test the whole program

Because `run` takes plain values, testing the full CLI needs no subprocesses:

{{< example file="projects/greet/cmd/greet/main_test.go" >}}

### Stamp the version

{{< example file="projects/greet/internal/version/version.go" >}}

### Automate builds with a Makefile

{{< example file="projects/greet/Makefile" >}}

```bash
cd projects/greet
make build
./bin/greet -version
```

```text
greet v0.1.0-3-g3f9c2ab (commit 3f9c2ab, built 2026-10-16T09:27:51Z)
```

Your version, commit and date will differ. Then cross-compile for four platforms in one command:

```bash
make release
ls dist
```

```text
greet-darwin-arm64  greet-linux-amd64  greet-linux-arm64  greet-windows-amd64.exe
```

{{% /steps %}}

## Exercise

The project uses `internal/greeter` with a slightly different API from the one below. This exercise asks you to write the parsing half yourself, in a form that is easy to test.

{{< exercise dir="m01/cli" >}}
Implement `Parse(args []string) (Config, error)` and `Render(cfg Config) string`. Use a `flag.FlagSet` with `flag.ContinueOnError`, validate that `-times` is at least 1, and return `ErrInvalidTimes` when it is not. The starter file contains detailed requirements.
{{< /exercise >}}

## Stretch goals

Once the tests pass, try these on your own copy of the project:

1. Add a `-lang` flag supporting `en`, `ta` and `hi` greetings. Return an error for unknown languages.
2. Read a default name from the `GREET_NAME` environment variable when `-name` is not given. Write a test for it using `t.Setenv`.
3. Install your build with `go install ./cmd/greet` and run `go version -m $(which greet)` to see the build info.

## Module 1 recap

You can now:

- Install Go, manage toolchains and set up your editor
- Create modules, add dependencies, and read `go.mod` and `go.sum`
- Organize code into packages with clear visibility and good names
- Build, cross-compile and stamp versions into binaries
- Keep code formatted, vetted and linted

Next up is **[Module 2: Language basics](/learn/02-language-basics/)**, where we go through Go's types, control flow, functions and pointers in detail.

## Further reading

- [Package flag](https://pkg.go.dev/flag)
- [Tutorial: Add a test](https://go.dev/doc/tutorial/add-a-test)
- [Organizing a Go module: commands](https://go.dev/doc/modules/layout#package-or-command-with-supporting-packages)
