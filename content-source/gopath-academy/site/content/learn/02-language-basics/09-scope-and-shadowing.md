---
title: "2.9 Scope and shadowing gotchas"
linkTitle: "2.9 Scope and shadowing"
weight: 9
description: "How Go's block scopes work, how := can silently shadow variables, and the tools that catch shadowing bugs."
goVersion: "1.27"
---

## What you will learn

- The levels of scope in Go, from universe to block
- What shadowing is and why `:=` makes it easy to do by accident
- The shadowed `err` bug, step by step
- How shadowing package names and built-ins causes confusing errors
- Tools that detect shadowing

## The idea

A **scope** is the region of code where a name refers to a particular declaration. Go has lexical (block) scoping, like most languages you know. The levels, from outermost to innermost:

| Scope | Contains | Example |
|---|---|---|
| Universe | Built-in names | `int`, `len`, `true`, `nil`, `error`, `make` |
| Package | Package-level declarations, across all files in the package | `var defaultPort`, `func Hello` |
| File | Imported package names | `fmt`, `strings` |
| Function | Parameters, results, local variables | `name`, `err` |
| Block | Anything declared inside `{ }`, and in `if`, `for` and `switch` init statements | `i` in `for i := ...` |

**Shadowing** happens when an inner scope declares a name that already exists in an outer scope. Inside the inner scope, the new declaration hides the outer one. Both variables exist; they are simply different variables with the same name.

## Shadowing in action

{{< example file="examples/m02/scope/scope_test.go" region="blocks" run="go test -run Example_blocks -v ./examples/m02/scope" >}}

Shadowing itself is not an error, and sometimes it is perfectly reasonable. The problem is that `:=` makes it **very easy to shadow by accident**, and the compiler stays quiet as long as both variables are used.

## The shadowed `err` bug

Almost every Go developer writes this bug once. The outer `err` is never set, because `:=` inside the `if` block declares a new `err`:

{{< example file="examples/m02/scope/scope_test.go" region="errshadow" run="go test -run Example_shadowedError -v ./examples/m02/scope" >}}

Walk through `buggy`:

1. `var err error` declares `err` in the function scope.
2. Inside the `if` block, `data, err := load(name)` needs a new variable (`data`), so `:=` declares **both** `data` and a **new** `err` in the block scope, as lesson 2.1's rules say. The "reuse existing variables" rule only applies to variables in the **same** scope.
3. The block ends and its `err` disappears.
4. `return err` returns the outer `err`, which is still `nil`.

The fix is to use `=` for variables that already exist, and declare the others separately if needed:

```go
var data string
data, err = load(name)
```

## Shadowing package names

Imported packages live in file scope, so a local variable can hide them:

```go
func fetch(url string) error {
	u, err := url.Parse(url) // url.Parse undefined (type string has no field or method Parse)
	...
}
```

The parameter `url` shadows the `net/url` package. The compiler also reports `"net/url" imported and not used`, because inside the function nothing can reach the package anymore. Rename the variable (`rawURL`, `target`) instead of fighting the error. Common collisions: `url`, `path`, `filepath`, `time`, `errors`, `json`, `log`, `user`.

## Shadowing built-ins

Built-in names are **not keywords**. They live in the universe scope and can be shadowed:

{{< example file="examples/m02/scope/scope_test.go" region="universe" run="go test -run Example_shadowBuiltins -v ./examples/m02/scope" >}}

Variables named `len`, `cap`, `copy`, `new`, `min`, `max`, `string` or `error` compile fine and then produce baffling errors a few lines later. Avoid these names entirely. (This is why the exercise in lesson 2.6 uses `lo` and `hi` rather than `min` and `max`.)

## Tools that catch it

- **Your editor.** `gopls` can highlight shadowed variables if you enable its `shadow` analyzer.
- **The `shadow` analyzer.** It is not part of the default `go vet` set because it reports some intentional shadowing, but you can run it directly:

```bash
go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
go vet -vettool=$(which shadow) ./...
```

- **golangci-lint** includes it through `govet` settings, and the `predeclared` linter flags shadowed built-ins.

The best prevention is a habit: inside a nested block, ask whether each `:=` is meant to create something new. If a variable already exists outside, use `=`.

## Common mistakes

### `:=` inside `if`, `for` or `switch` blocks for an outer variable

The `err` bug above. It also happens with results you meant to collect, like `total := total + n` inside a loop (that creates a new `total` each iteration).

### Naming variables after imported packages

`url`, `json`, `time`, `errors`. Pick names that describe the value instead.

### Naming variables after built-ins

`len`, `max`, `string`, `copy`. They compile, then break code further down.

### Assuming shadowing is always wrong

The `if v, ok := m[k]; ok { }` idiom and `err` inside a short `if err := f(); err != nil` are scoped declarations that do not shadow anything important. The goal is to avoid **accidental** shadowing, not to forbid nested scopes.

## Key takeaways

- Go has universe, package, file, function and block scopes.
- An inner declaration with the same name **shadows** the outer one; both variables exist.
- `:=` in a nested block always declares new variables, which causes the classic shadowed `err` bug.
- Avoid naming variables after imported packages or built-ins.
- Use the `shadow` analyzer or linters to catch accidental shadowing.

## Further reading

- [Declarations and scope](https://go.dev/ref/spec#Declarations_and_scope) in the spec
- [Blocks](https://go.dev/ref/spec#Blocks) in the spec
- [shadow analyzer](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/shadow)
