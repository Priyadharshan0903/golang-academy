---
title: "1.5 Packages, visibility and project layout"
linkTitle: "1.5 Packages and layout"
weight: 5
description: "How Go packages work, how exported names replace public and private, how to name packages, and how to lay out a project without over-engineering it."
goVersion: "1.27"
---

## What you will learn

- What a package is, and how import paths map to directories
- How Go decides what is visible outside a package
- The special `internal/` and `cmd/` folders
- How to name packages well
- A sensible project layout, and what not to copy from other languages

## The idea

A **package** is a directory of `.go` files that all start with the same `package` clause. The files are compiled together and can freely use each other's names, even unexported ones. One directory, one package.

A package's **import path** is the module path plus the directory:

| Module path | Directory | Import path |
|---|---|---|
| `github.com/you/shop` | `cart/` | `github.com/you/shop/cart` |
| `github.com/you/shop` | `internal/db/` | `github.com/you/shop/internal/db` |

## Visibility: the capital letter rule

Go has no `public`, `private` or `protected` keywords. Visibility is decided by the **first letter of the name**:

- Starts with an uppercase letter: **exported**, usable from other packages.
- Anything else: **unexported**, visible only inside the same package.

This applies to everything: functions, types, struct fields, methods, constants and variables.

{{< example file="examples/m01/greeting/greeting.go" region="visibility" >}}

From another package, `greeting.Hello` and `greeting.DefaultName` are available, and `greeting.normalize` does not exist. The test below lives in the external test package `greeting_test`, so it sees the package exactly as a user would:

{{< example file="examples/m01/greeting/greeting_test.go" region="external" run="go test ./examples/m01/greeting" >}}

A test file with the **same** package name sits "inside" the package and can test unexported helpers directly:

{{< example file="examples/m01/greeting/normalize_test.go" region="internal" >}}

Visibility is per **package**, not per file or per type. Two files in the same package see each other's unexported names, and there is no way to hide a field from other code in the same package. This pushes you toward small, focused packages.

## Package names

The package name is what callers type, so it shapes how your API reads: `greeting.Hello("Ana")`.

Good package names are:

- **Short, lowercase, one word**: `http`, `json`, `cart`, `auth`. No underscores, no `mixedCaps`.
- **The last element of the import path**: the package in `.../internal/ratelimit` should be `package ratelimit`.
- **About what it provides**, not what it contains: `user` instead of `models`, `httputil` instead of `helpers`.

Avoid **stutter**. Because callers always write the package name, repeating it in identifiers is redundant:

| Stutters | Reads well |
|---|---|
| `user.UserService` | `user.Service` |
| `cart.NewCart()` | `cart.New()` |
| `config.ConfigLoad()` | `config.Load()` |

Avoid grab-bag names like `util`, `common`, `misc` and `base`. They give no hint about their contents and tend to collect unrelated code until everything depends on them.

## Special folders

### `internal/`

A package under a directory named `internal` can only be imported by code rooted at the parent of that `internal` directory. The compiler enforces this.

```text
github.com/you/shop/internal/db      importable by anything under github.com/you/shop/
                                      NOT importable by any other module
```

Use `internal/` for everything you do not intend to support as public API. You can then change it freely without breaking anyone.

### `cmd/`

By convention, each executable lives in its own folder under `cmd/`, named after the binary:

```text
cmd/greet/main.go    ->  go build ./cmd/greet  produces "greet"
cmd/greetd/main.go   ->  go build ./cmd/greetd produces "greetd"
```

Keep `main` packages thin: parse flags, wire things together, call into `internal/` packages that hold the real logic.

### `testdata/`

The `go` tool ignores directories named `testdata`. Put fixture files for tests there.

## A sensible layout

This is the layout of the checkpoint project you will build at the end of this module:

{{< filetree/container >}}
  {{< filetree/folder name="projects/greet" >}}
    {{< filetree/folder name="cmd" >}}
      {{< filetree/folder name="greet" >}}
        {{< filetree/file name="main.go" >}}
        {{< filetree/file name="main_test.go" >}}
      {{< /filetree/folder >}}
    {{< /filetree/folder >}}
    {{< filetree/folder name="internal" >}}
      {{< filetree/folder name="greeter" >}}
        {{< filetree/file name="greeter.go" >}}
        {{< filetree/file name="greeter_test.go" >}}
      {{< /filetree/folder >}}
      {{< filetree/folder name="version" >}}
        {{< filetree/file name="version.go" >}}
        {{< filetree/file name="version_test.go" >}}
      {{< /filetree/folder >}}
    {{< /filetree/folder >}}
    {{< filetree/file name="Makefile" >}}
    {{< filetree/file name="README.md" >}}
  {{< /filetree/folder >}}
{{< /filetree/container >}}

And the right amount of structure depends on size:

- **A small tool or library**: everything in the root package. One folder, a few files. That is fine and common.
- **An app with one binary**: `main.go` at the root or in `cmd/name/`, logic in a few packages under `internal/`.
- **Several binaries**: one folder per binary in `cmd/`, shared code in `internal/`.

Start flat and add packages when a clear boundary appears. Splitting too early creates import cycles and packages that only exist to hold one type.

{{< callout type="warning" >}}
You will find a popular repository called "golang-standards/project-layout". Despite the name, it is **not** an official standard and is not maintained by the Go team. The official guidance is [Organizing a Go module](https://go.dev/doc/modules/layout), and it is much simpler.
{{< /callout >}}

## Initialization order, briefly

When a program starts, Go initializes packages from the bottom of the import graph upward: a package is fully initialized before any package that imports it.

Within a package, package-level variables are initialized first (in dependency order), then any `init()` functions run in the order they appear. Only then does `main()` start.

```go
var startedAt = time.Now()

func init() {
	// runs once, before main, after startedAt is set
}
```

Use `init()` sparingly. Hidden work at import time makes code harder to test and reason about. Prefer explicit setup functions that `main` calls.

## Common mistakes

### Import cycles

Package `a` imports `b`, and `b` imports `a`. Go forbids this and fails with `import cycle not allowed`. The fix is almost always a better boundary: move the shared piece into a third package, or have one side depend on a small interface instead of the concrete package.

### One package per type, Java style

Creating `models/user`, `models/order`, `services/user`, `services/order` and so on spreads one concept across many packages and makes cycles likely. In Go, group by what code is **about** (`user`, `order`), and keep related types together.

### Exporting everything

Uppercase names are a promise to callers. Export only what users of the package need. You can always export something later; un-exporting breaks code.

### A package name that does not match its folder

Legal, but confusing: readers see `import ".../storage"` and then `db.Open()`. Keep them identical.

## Exercise

{{< exercise dir="m01/tempconv" >}}
Implement `CToF`, `FToC` and `IsPhysical` in the `tempconv` package. The tests are in an **external** test package (`tempconv_test`), so they can only call exported names. Notice how that forces you to think about the package from a caller's point of view.
{{< /exercise >}}

## Key takeaways

- One directory is one package; the import path is the module path plus the directory.
- Uppercase first letter means exported. Visibility is per package.
- Name packages with short, lowercase nouns and avoid stutter and grab-bag names.
- `internal/` hides packages from other modules; `cmd/` holds one folder per binary.
- Start with a flat layout and add packages only when a real boundary appears.

## Further reading

- [Organizing a Go module](https://go.dev/doc/modules/layout), the official layout guide
- [Package names](https://go.dev/blog/package-names) on the Go blog
- [Effective Go: Names](https://go.dev/doc/effective_go#names)
- [Package initialization](https://go.dev/ref/spec#Package_initialization) in the spec
