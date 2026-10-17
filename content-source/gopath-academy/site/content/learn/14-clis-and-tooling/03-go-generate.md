---
title: "14.3 Code generation with go generate"
linkTitle: "14.3 go generate"
weight: 3
description: "Generate Go code instead of writing it by hand: go:generate directives, running tools pinned with the go.mod tool directive (Go 1.24), writing your own generator with go/types and text/template, and keeping generated code honest in CI."
goVersion: "1.27"
---

## What you will learn

- What `go generate` does, and what it deliberately does not do
- Running a pinned generator with `go tool` and the **`tool` directive** (Go 1.24)
- Writing your own generator with `go/parser`, `go/types` and `text/template`
- The conventions for generated files
- Making sure generated code never goes stale

## The idea

Some code is mechanical: `String` methods for enum constants, mocks for interfaces, Go types from a protobuf or OpenAPI schema, database access code from SQL queries. Writing it by hand is tedious and error-prone; generating it means changing one source and re-running a tool.

`go generate` is the standard way to run such tools. It scans Go files for comments like:

```go
//go:generate go tool stringer -type=Level -trimprefix=Level
```

and runs each command in the directory of the file that contains it. That is all it does. It is **not** part of `go build`: you run it yourself when the inputs change, and you **commit the generated files**. This keeps builds simple and fast, and means people using your package never need your generators installed.

```bash
go generate ./...     # run every directive in the module
```

## Pinned tools: the tool directive (Go 1.24)

Generators are programs, and everyone on a team should run the same version. Since Go 1.24, a module can declare tools in `go.mod`:

```bash
go get -tool golang.org/x/tools/cmd/stringer
```

adds a line to `go.mod`, and the version to the `require` block:

```text
tool golang.org/x/tools/cmd/stringer
```

Now `go tool stringer` builds and runs exactly that version, with no separate installation and no `tools.go` file with blank imports (the old workaround). The `examples/thirdparty` module does this:

{{< example file="examples/thirdparty/genstringer/level.go" >}}

Running `go generate` there writes `level_string.go`:

{{< example file="examples/thirdparty/genstringer/level_string.go" >}}

Two details of `stringer`'s output are worth noticing. The strange `func _()` block is a **compile-time check**: if someone renumbers the constants without regenerating, an array index goes out of range and the build fails. And all names are packed into one string with an index table, avoiding a `switch` and allocations.

## Writing your own generator

When no existing tool fits, writing one is easier than it sounds, because the standard library ships the Go parser and type checker. This generator produces `String`, `Parse`, `Values` and text marshaling methods for an enum type:

{{< example file="examples/m14/color/color.go" >}}

The generator is a normal Go program in the same module, run with `go run`:

{{< example file="examples/m14/enumgen/main.go" region="generate" >}}

The steps every generator follows:

1. **Parse** the package's files with `go/parser` (skipping tests and its own output).
2. **Type-check** them with `go/types`. This is what makes the generator reliable: the type checker evaluates `iota` and constant expressions, so `ColorRed Color = iota + 1` is known to be 1 without the generator interpreting anything itself.
3. **Collect** the facts it needs: here, every constant whose type is `Color`.
4. **Render** code with `text/template`:

{{< example file="examples/m14/enumgen/main.go" region="template" >}}

5. **Format** with `go/format`, so the output is always `gofmt`-clean.

The result is used like hand-written code:

{{< example file="examples/m14/color/color_test.go" region="example" run="go test -v ./examples/m14/color" >}}

Because the generated type implements `encoding.TextMarshaler`, JSON encodes it as `"red"` rather than `1`; the same methods also make it work with `flag.TextVar` and in `slog` output.

## Conventions for generated files

- Start the file with a comment matching `^// Code generated .* DO NOT EDIT\.$` (the convention is documented in `go help generate`). Tools recognize it: linters such as golangci-lint skip such files by default, coverage and review tools can leave them out, and GitHub collapses them in pull request diffs. Both generators above follow it.
- Name the file after its source and generator: `level_string.go`, `color_enum.go`.
- Put the `//go:generate` line next to the declaration it is about.
- Never edit generated files by hand. Change the input or the generator, and regenerate.

## Keeping generated code honest

The risk with committed generated code is that someone changes the input and forgets to regenerate. Two defenses, both used in this course:

{{< example file="examples/m14/enumgen/main_test.go" region="uptodate" >}}

A **test** that runs the generator in memory and compares with the committed file. And a **CI step** that regenerates everything and fails if anything changed:

```yaml
- name: Generated code is up to date
  run: |
    go generate ./...
    git diff --exit-code
```

## Generators you will meet

| Tool | Generates |
|---|---|
| `stringer` (golang.org/x/tools) | `String` methods for constants |
| `protoc-gen-go`, `buf` | Go types and gRPC code from protobuf |
| `oapi-codegen`, `ogen` | servers and clients from OpenAPI |
| `sqlc` | type-safe database code from SQL queries |
| `mockgen`, `moq` | mocks from interfaces |
| `controller-gen` | Kubernetes CRDs and deep-copy methods (Module 15) |

## When not to generate

Generics (Module 4) removed the most common reason for generators before Go 1.18: writing the same container or algorithm for several types. Prefer a generic function to generated copies. And `//go:embed` (lesson 8.5) is the right tool for including static files; do not generate Go source from them.

## Common mistakes

### Expecting go build to run generators

It never does. Run `go generate` and commit the output.

### Unpinned generator versions

Different developers produce different output, and diffs churn. Use the `tool` directive.

### Editing generated files

The next `go generate` erases the change.

### Generated code that is not gofmt-clean

Run the output through `go/format.Source`.

## Key takeaways

- `go generate` runs `//go:generate` commands; it is a developer tool, not part of the build. Commit generated files.
- Pin generators with `go get -tool` and run them with `go tool name` (Go 1.24).
- Write generators with `go/parser` + `go/types` (to understand the code) and `text/template` + `go/format` (to write it).
- Mark files with `// Code generated ... DO NOT EDIT.`, and check freshness with a test or `git diff --exit-code` in CI.

## Further reading

- [Generating code](https://go.dev/blog/generate) on the Go blog
- [go.mod reference: tool directive](https://go.dev/ref/mod#go-mod-file-tool)
- [Package go/types](https://pkg.go.dev/go/types), and Alan Donovan's [go/types tutorial](https://github.com/golang/example/tree/master/gotypes)
