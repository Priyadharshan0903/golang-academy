---
title: "14.1 flag vs Cobra"
linkTitle: "14.1 flag vs Cobra"
weight: 1
description: "Build command-line tools with subcommands using the standard flag package, handle files and standard input, use exit codes properly, make the whole program testable, and see what Cobra adds."
goVersion: "1.27"
---

## What you will learn

- The conventions users expect from a command-line tool
- Subcommands with the standard `flag` package, one `FlagSet` per command
- Reading files or standard input, writing to stdout and stderr
- Exit codes, and how Ctrl+C should behave
- Testing a CLI completely, in-process
- What **Cobra** gives you, and when it is worth the dependency

## Conventions worth following

Unix tools agree on a few rules, and tools that follow them compose with everything else:

- **Results go to stdout, messages to stderr.** Then `tool | jq` works, and errors still reach the terminal.
- **Read the files named as arguments, or standard input if there are none.** Then `tool file.json` and `cat file.json | tool` both work.
- **Exit code 0 means success.** 1 is a general error; 2 is a usage error (the `flag` package exits with 2). Scripts check exit codes, not messages.
- **`-h` prints help**, and is not an error.
- **Ctrl+C stops promptly** and cleanly.

## The example: jsonq

`jsonq` is a small tool with three subcommands, built with only the standard library:

```bash
$ echo '{"name":"Ada","emails":["ada@example.com","countess@example.org"]}' | jsonq get -raw emails.1
countess@example.org
$ jsonq get -h
usage: jsonq get [flags] PATH [FILE...]
  -raw
    	print strings without quotes
```

### main is three lines

{{< example file="examples/m14/jsonq/main.go" >}}

`main` connects the program to the real process: a context cancelled by Ctrl+C, the real arguments and streams, and the exit code. Everything else is in `run`, which takes all of these as parameters. This is the same pattern as the greet CLI (lesson 1.8) and the shortener's `run` (lesson 12.8), and it is what makes the program testable.

### Subcommands

{{< example file="examples/m14/jsonq/run.go" region="run" >}}

The `flag` package has no notion of subcommands, but you do not need one: look at `args[0]`, and give each command its own **`flag.FlagSet`**. A table of commands keeps dispatch and the usage message in one place.

Errors are classified once, at the top: wrapping `errUsage` marks errors that are the caller's fault (exit 2), anything else is a runtime error (exit 1), and `flag.ErrHelp` (the user asked for `-h`) is success.

{{< example file="examples/m14/jsonq/run.go" region="flagset" >}}

`flag.ContinueOnError` makes `Parse` **return** errors instead of calling `os.Exit(2)` itself, which is what a testable program needs. `SetOutput` sends the flag package's messages to the command's stderr, and a custom `Usage` adds the positional arguments to the help text.

Note a limitation of `flag`: it stops parsing at the first non-flag argument. `jsonq get emails.1 -raw` treats `-raw` as a file name; flags must come first. Cobra (below) relaxes this.

### Files or standard input

{{< example file="examples/m14/jsonq/run.go" region="input" >}}

Reading through an `io.Reader` makes files and stdin the same thing (lesson 8.1). A `json.Decoder` reads value after value, so files with several JSON documents (JSON Lines) work for free. Checking `ctx.Err()` between documents makes Ctrl+C stop the tool cleanly in the middle of a large input.

## Testing the whole CLI

{{< example file="examples/m14/jsonq/run_test.go" region="example" run="go test -v -run Example ./examples/m14/jsonq" >}}

Because `run` takes its arguments, input and outputs as parameters, a test can execute the whole program in-process, many times per millisecond, and check exactly what a user would see: stdout, stderr and the exit code. No subprocess, no build step. The other tests cover help, unknown commands, missing files and cancellation.

## Cobra

`github.com/spf13/cobra` is the library behind `kubectl`, `docker`, `gh`, `hugo` and many more. Here are two of jsonq's commands in Cobra:

{{< example file="examples/thirdparty/clicobra/cli.go" region="root" >}}

What you get, compared with `flag`:

- **A command tree**, with nested subcommands (`kubectl config set-context ...`).
- **POSIX-style flags** through `pflag`: `--long` and `-s` short forms, and flags allowed after positional arguments.
- **Persistent flags** inherited by subcommands (`--compact` here, `--namespace` in kubectl).
- **Argument validation** (`cobra.ExactArgs(1)`), and generated help and usage:

```text
Answer simple questions about JSON

Usage:
  jsonq [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  get         Print the value of a top-level key
  help        Help about any command
  keys        Print the top-level keys of an object read from stdin

Flags:
      --compact   print JSON without indentation
  -h, --help      help for jsonq
```

- **Shell completion** for bash, zsh, fish and PowerShell, generated: `jsonq completion bash`.

Two habits keep Cobra code testable. Build the tree in a **function** (`NewRootCmd`) instead of package-level variables wired up in `init`, so each test gets a fresh tree. And use `cmd.InOrStdin()`, `cmd.OutOrStdout()` and `RunE` (returning errors) rather than the global `os.Stdin`, `fmt.Println` and `os.Exit`:

{{< example file="examples/thirdparty/clicobra/cli_test.go" region="test" run="cd examples/thirdparty && go test -v -run Example ./clicobra" >}}

## Which to use

Use **`flag`** for tools with a handful of flags and up to a few subcommands: no dependencies, and everything is visible in a page of code. Most internal tools and many open source ones never need more. Use **Cobra** when the tool grows a real command hierarchy, when users expect GNU-style flags and completion, or when you are building a `kubectl` plugin or another tool in that ecosystem. (`urfave/cli` and `kong` are popular alternatives with different styles.)

## Common mistakes

### Calling os.Exit or log.Fatal deep in the code

Deferred functions do not run, and the code becomes untestable. Return errors to `run`; exit only in `main`.

### Printing errors to stdout

They end up in the next program of a pipeline. Errors go to stderr.

### Exit code 0 on failure

Scripts and CI then believe it worked. Return non-zero on any failure.

### Global state in Cobra commands

Package-level flag variables leak between tests. Build commands in functions.

## Key takeaways

- Results to stdout, messages to stderr; files as arguments or stdin; exit 0, 1 or 2.
- Keep `main` tiny; put the program in `run(ctx, args, stdin, stdout, stderr) int` and test it in-process.
- With `flag`, use one `FlagSet` per subcommand with `ContinueOnError`; classify errors once at the top.
- Cobra adds command trees, POSIX flags, persistent flags, generated help and shell completion; build its commands in functions and use `RunE` and the command's streams.

## Further reading

- [Package flag](https://pkg.go.dev/flag)
- [Cobra user guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md)
- [Command Line Interface Guidelines](https://clig.dev/)
