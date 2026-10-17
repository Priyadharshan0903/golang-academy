---
title: "1.3 Your first program, line by line"
linkTitle: "1.3 First program"
weight: 3
description: "Write, run and build a Go program, understand every line of it, and learn the compiler rules that surprise newcomers."
goVersion: "1.27"
---

## What you will learn

- What every line of a minimal Go program does
- The difference between `go run` and `go build`
- The strict compiler rules that catch experienced developers off guard
- How this course uses testable examples to keep code correct

## The program

{{< example file="examples/m01/hello/main.go" run="go run ./examples/m01/hello" >}}

```text
Hello, Gopher!
```

Only a few lines, and each one carries a rule worth knowing.

### `package main`

Every Go file starts by declaring which package it belongs to. A package is Go's unit of code organization: a directory of `.go` files that are compiled together.

The name `main` is special. It tells the compiler to build an **executable** instead of a library. Any other package name produces a library that other code imports.

### `import "fmt"`

`fmt` is the standard library package for formatted input and output. Imports are by path: `"fmt"` for the standard library, or something like `"github.com/google/uuid"` for a third party package.

Several imports are grouped in parentheses:

```go
import (
	"fmt"
	"os"
)
```

### `func main()`

In package `main`, the function `main` is the entry point. It takes no arguments and returns nothing. Command line arguments come from `os.Args`, and you set the exit code with `os.Exit`. We use both in the checkpoint project.

### `fmt.Println("Hello, Gopher!")`

You call a function from another package as `package.Function`. `Println` is capitalized, which is how Go marks a name as **exported** (public). Lowercase names are private to their package. That one rule replaces `public` and `private` keywords entirely, and lesson 1.5 covers it in depth.

## Run it, build it

Create a new folder anywhere and try it yourself:

```bash
mkdir hello && cd hello
go mod init example.com/hello
```

Save the program as `main.go`, then:

```bash
go run .
```

`go run` compiles the program to a temporary location and runs it immediately. It is for quick iteration.

```bash
go build -o hello .
./hello
```

`go build` produces a real executable in the current folder. There is nothing else to ship: no virtual machine, no interpreter, no `node_modules`. You can copy `hello` to another machine with the same operating system and architecture and run it.

{{< callout type="info" >}}
Use `go run .` (a dot, meaning "the package in this directory") instead of `go run main.go`. As soon as your program has a second file, `go run main.go` fails because it only compiles the files you list.
{{< /callout >}}

## The compiler is strict, on purpose

Go refuses to compile some things that other languages only warn about. This feels annoying for a week and helpful forever after.

### Unused imports and variables are errors

```go
import "os" // error: "os" imported and not used
```

```go
x := 42 // error: declared and not used: x
```

Unused code in a large codebase is noise and often a sign of a bug. Go makes it a hard error. Your editor's "organize imports" action (from lesson 1.2) adds and removes imports for you.

When you really want to ignore a value, assign it to the blank identifier `_`:

```go
_, err := fmt.Println("hi") // ignore the byte count, keep the error
```

### The opening brace must be on the same line

```go
func main()
{ // syntax error: unexpected semicolon or newline before {
```

Go's grammar uses semicolons, but you never type them. The lexer inserts one automatically at the end of a line that ends with an identifier, a literal, `)` or `}`. A line ending in `func main()` gets a semicolon, which breaks the declaration. Keeping `{` on the same line avoids that, and `gofmt` always formats code this way.

## How this course checks its own examples

The program above is a real file in the course repository, and it is tested. Next to it is this test file:

{{< example file="examples/m01/hello/main_test.go" >}}

A function named `Example` (or `ExampleSomething`) in a `_test.go` file is a **testable example**. `go test` runs it, captures what it prints, and compares that with the `// Output:` comment. If they differ, the test fails.

```bash
go test -v ./examples/m01/hello
```

```text
=== RUN   Example
--- PASS: Example (0.00s)
PASS
ok  	github.com/Priyadharshan0903/gopath-academy/examples/m01/hello	0.003s
```

Every example in this course works this way, so the output shown in a lesson is exactly what the code prints. You will write your own examples and tests in Module 9.

## Common mistakes

### Running a single file

`go run main.go` works until you split your program into two files, then fails with `undefined:` errors for everything in the other file. Use `go run .`.

### Fighting unused-variable errors with dummy assignments

Writing `_ = x` everywhere to silence the compiler hides real bugs. Delete the variable, or finish the code that was supposed to use it.

### Expecting a `main` package to be importable

You cannot import a `package main` from other code. Put reusable logic in a separate package and keep `main` thin. The checkpoint project shows this pattern.

## Exercise

{{< exercise dir="m01/greeting" >}}
Implement `Greeting(name string) string` so it returns `Hello, <name>!`. Ignore leading and trailing spaces, and use `Gopher` when the name is empty. Look at [`strings.TrimSpace`](https://pkg.go.dev/strings#TrimSpace).
{{< /exercise >}}

## Key takeaways

- `package main` plus `func main()` makes an executable; every other package is a library.
- Names starting with an uppercase letter are exported. That is Go's whole visibility system.
- `go run .` for quick iteration, `go build` for a real binary.
- Unused imports and variables are compile errors, and the opening brace must stay on the same line.
- Testable examples with `// Output:` keep documentation and code in sync.

## Further reading

- [Tutorial: Get started with Go](https://go.dev/doc/tutorial/getting-started)
- [A Tour of Go](https://go.dev/tour/), great for quick practice alongside this course
- [Semicolons](https://go.dev/ref/spec#Semicolons) in the language specification
- [Testable examples in Go](https://go.dev/blog/examples) on the Go blog
