---
title: "1.7 Formatting, vetting and linting"
linkTitle: "1.7 Format, vet, lint"
weight: 7
description: "gofmt, goimports, go vet, staticcheck and golangci-lint: what each catches, and how to wire them into your editor and CI."
goVersion: "1.27"
---

## What you will learn

- Why Go has exactly one code style, and how `gofmt` enforces it
- What `go vet` catches that the compiler does not
- When to add `staticcheck` or `golangci-lint`
- How to write doc comments that tools understand
- A minimal setup for your editor and CI

## The idea

Go treats code quality tooling as part of the language experience, not an afterthought. Three layers, from strictest to most opinionated:

1. **The compiler** rejects code that is wrong (unused variables, type errors).
2. **`gofmt` and `go vet`** ship with Go. Formatting is non-negotiable, and vet flags code that compiles but is probably a bug.
3. **Linters** like `staticcheck` and `golangci-lint` add deeper checks and style rules. You choose how many.

## `gofmt`: one style for everyone

`gofmt` rewrites Go source into the canonical format: tabs for indentation, aligned comments and struct fields, consistent spacing and brace placement. There are no options to argue about.

```bash
gofmt -l .        # list files that are not formatted
gofmt -d file.go  # show the diff it would apply
gofmt -w .        # rewrite files in place
go fmt ./...      # same as gofmt -l -w on every package
```

The result is that all Go code, in every company and every open source project, looks the same. Code review is about logic, not style, and reading unfamiliar code is easier.

### `goimports`

`goimports` does everything `gofmt` does, plus adds missing imports and removes unused ones. Most editors run it (or the equivalent `gopls` action) on save, which is why you rarely type import lines by hand.

```bash
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .
```

## `go vet`: compiles, but probably wrong

`go vet` runs a set of analyzers that find suspicious code. A classic example is a `Printf` format that does not match its arguments:

```go
fmt.Printf("%d items\n", "three") // compiles fine
```

```bash
go vet ./...
```

```text
./main.go:8:2: fmt.Printf format %d has arg "three" of wrong type string
```

Other things vet catches include copying a `sync.Mutex` by value, malformed struct tags, unreachable code, invalid `errors.As` targets and suspicious shifts. The checks are chosen to have very few false positives, so treat every vet warning as a bug until proven otherwise.

`go test` runs a subset of the vet checks automatically before running tests, which is why you sometimes see vet errors from `go test`.

## `go fix`: modernizing code

Go 1.26 rewrote `go fix` as a home for **modernizers**: automated rewrites that update old patterns to newer, simpler forms (for example, replacing hand-written loops with functions from the `slices` package). It only applies changes that keep behavior the same.

```bash
go fix -diff ./...   # preview
go fix ./...         # apply
```

Run it after upgrading Go, review the diff, and commit it separately from other changes.

## Linters

### `staticcheck`

[staticcheck](https://staticcheck.dev) is the most widely used standalone linter. It finds bugs, performance problems, deprecated API usage and simplifications, with a strong focus on avoiding false positives.

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

If you only add one linter, add this one.

### `golangci-lint`

[golangci-lint](https://golangci-lint.run) runs many linters (including staticcheck's checks) in one fast pass with a single config file. It is the common choice in CI for teams.

```bash
golangci-lint run ./...
```

A conservative starting config, `.golangci.yml`:

```yaml
version: "2"
linters:
  default: standard
  enable:
    - errorlint
    - misspell
    - revive
```

Start small. A config that enables every available linter produces hundreds of warnings, people learn to ignore them, and the tool stops being useful.

{{< callout type="info" >}}
golangci-lint recommends installing a pinned release binary rather than using `go install`, so that everyone gets the same build. Follow the [install guide](https://golangci-lint.run/welcome/install/) for your platform. You can also pin it with the `tool` directive from lesson 1.4.
{{< /callout >}}

## Doc comments

A comment directly above a declaration, with no blank line in between, is its **doc comment**. `go doc`, `gopls` hovers and [pkg.go.dev](https://pkg.go.dev) all display it.

Conventions:

- Start with the name of the thing being documented: `// Hello returns ...`.
- Write full sentences.
- Every exported name deserves a doc comment. Linters like `revive` can enforce it.
- The **package comment** goes above the `package` clause in one file, starting with `// Package name ...`.

{{< example file="examples/m01/greeting/greeting.go" region="doc" >}}

```bash
go doc ./examples/m01/greeting
```

```text
package greeting // import "github.com/Priyadharshan0903/gopath-academy/examples/m01/greeting"

Package greeting builds friendly greetings.

It exists to show how Go decides what is visible outside a package: names that
start with an uppercase letter are exported, everything else is private to the
package.

const DefaultName = "Gopher"
func Hello(name string) string
```

## A minimal quality setup

**In your editor**: format on save and organize imports on save (lesson 1.2). Most editors also show vet and staticcheck results inline through `gopls`.

**In CI**, fail the build if any of these fail:

```bash
test -z "$(gofmt -l .)"   # formatting
go vet ./...              # vet
go test ./...             # tests
staticcheck ./...         # or: golangci-lint run
```

This course's own CI does the same. You can see it in [`.github/workflows/ci.yml`](https://github.com/Priyadharshan0903/gopath-academy/blob/main/.github/workflows/ci.yml).

## Common mistakes

### Configuring your editor to format differently

Some editors default to spaces or a different line width. Let `gofmt` (or `goimports`) be the formatter. Never fight it.

### Silencing vet instead of fixing the code

Vet warnings are rarely false positives. If vet complains about a copied mutex, the program really has two independent locks where you think there is one.

### Turning on every linter at once

Large rule sets on an existing codebase create noise that hides the warnings that matter. Begin with `staticcheck` or golangci-lint's default set, fix what it finds, and add linters one at a time.

### Doc comments that repeat the signature

`// GetUser gets the user` adds nothing. Say what callers need to know: what happens when the user does not exist, whether it is safe for concurrent use, what units a value is in.

## Key takeaways

- `gofmt` gives all Go code one style. Format on save and stop thinking about it.
- `go vet` catches real bugs that compile. Run it everywhere; `go test` runs part of it automatically.
- `go fix` applies safe modernizations after upgrades.
- Add `staticcheck` first, or `golangci-lint` with a small config, and grow from there.
- Doc comments start with the name they document and appear in `go doc`, editors and pkg.go.dev.

## Further reading

- [gofmt](https://pkg.go.dev/cmd/gofmt) and [Effective Go: Formatting](https://go.dev/doc/effective_go#formatting)
- [cmd/vet](https://pkg.go.dev/cmd/vet), list of analyzers
- [Go Doc Comments](https://go.dev/doc/comment), the full guide to doc comment syntax
- [staticcheck documentation](https://staticcheck.dev/docs/)
- [golangci-lint configuration](https://golangci-lint.run/usage/configuration/)
