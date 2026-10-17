---
title: "1.4 Go modules: go.mod, go.sum and dependencies"
linkTitle: "1.4 Go modules"
weight: 4
description: "How Go modules define a project, record dependencies, pick versions and verify downloads."
goVersion: "1.27"
---

## What you will learn

- What a module is, and how it relates to packages
- How to read every line of a `go.mod` file
- What `go.sum` is for, and why you commit it
- The everyday commands: `go mod init`, `go get`, `go mod tidy`
- How Go picks dependency versions, and the `/v2` rule for major versions

## The idea

A **module** is a collection of packages that are released and versioned together. A module is defined by a `go.mod` file at its root. Everything in that directory tree (except folders that contain their own `go.mod`) belongs to the module.

- **Package**: a directory of `.go` files. The unit you import.
- **Module**: a tree of packages with one `go.mod`. The unit you version and download.

Most projects are a single module. This course repository, for example, has one module at the root for examples and projects, and a separate module in `exercises/`.

## Creating a module

```bash
mkdir weather && cd weather
go mod init github.com/yourname/weather
```

```text
go: creating new go.mod: module github.com/yourname/weather
```

The argument is the **module path**. It is the prefix of every import path in your module, so a package in `weather/forecast/` is imported as `github.com/yourname/weather/forecast`.

Use the repository URL as the module path if you will publish the code. For code that will never be imported by anyone else, any name works, but a domain-like path (`example.com/weather`) avoids clashing with standard library names.

## Reading `go.mod`

A real `go.mod` after adding a few dependencies:

```text
module github.com/yourname/weather

go 1.27

require (
	github.com/google/uuid v1.6.0
	golang.org/x/sync v0.10.0
)

require golang.org/x/text v0.21.0 // indirect

tool golang.org/x/tools/cmd/stringer
```

| Directive | Meaning |
|---|---|
| `module` | The module path. |
| `go` | The minimum Go version this module needs. Also enables language features of that version. |
| `toolchain` | (Optional) a preferred toolchain, for example `toolchain go1.27.1`. |
| `require` | Dependencies and the **minimum** version of each. |
| `// indirect` | A dependency you do not import directly; one of your dependencies needs it. |
| `tool` | (Go 1.24+) a command line tool the project uses, run with `go tool stringer`. |
| `replace` | Use a different version or a local copy of a module. Only applies in the main module. |
| `exclude` | Never use a specific version of a module. |
| `retract` | Mark versions of **your own** module as broken so users avoid them. |

The version numbers above are illustrative. Your file will show whatever versions `go get` resolves.

## Adding and updating dependencies

```bash
# Add a dependency, or change its version
go get github.com/google/uuid@v1.6.0
go get github.com/google/uuid@latest

# Update all direct and indirect dependencies to newer minor or patch versions
go get -u ./...

# Remove a dependency (or just delete the import and run tidy)
go get github.com/google/uuid@none
```

In practice, the most common workflow is simpler: add an `import` to your code, then run:

```bash
go mod tidy
```

`go mod tidy` adds missing requirements, removes unused ones, and updates `go.sum`. Run it before every commit that touches imports.

### Tools as dependencies

Go 1.24 added the `tool` directive, so developer tools are versioned in `go.mod` like any other dependency:

```bash
go get -tool golang.org/x/tools/cmd/stringer
go tool stringer -type=Color
```

Everyone on the team, and CI, runs exactly the same version of the tool.

## `go.sum`: verifying downloads

`go.sum` contains cryptographic hashes of every module version your build uses:

```text
github.com/google/uuid v1.6.0 h1:<hash of the module's files>
github.com/google/uuid v1.6.0/go.mod h1:<hash of its go.mod file>
```

Each module version usually has two lines: one hash for the full source tree, and one for just its `go.mod` (which is all Go needs when it only has to read the dependency's requirements).

When the `go` command downloads a module, it checks the hash against `go.sum`. If someone tampered with the code, or a tag was moved, the build fails.

- **Commit `go.sum`** alongside `go.mod`.
- **Never edit it by hand.** `go mod tidy` maintains it.
- Merge conflicts in `go.sum` are resolved by taking both sides and running `go mod tidy`.

## Where modules come from

By default the `go` command downloads modules through the public Go module proxy and verifies them against the public checksum database:

```bash
go env GOPROXY GOSUMDB
```

```text
https://proxy.golang.org,direct
sum.golang.org
```

For private code (your company's repositories), tell Go not to use the public proxy or checksum database for those paths:

```bash
go env -w GOPRIVATE=github.com/yourcompany/*
```

## How versions are chosen

Go modules use **semantic versioning**: `vMAJOR.MINOR.PATCH`. A minor or patch release must stay backwards compatible. A new major version may break things.

Go then picks versions with **minimal version selection (MVS)**. For each dependency, it uses the **highest of the minimum versions** that anything in your build asks for. It never silently picks a newer version just because one exists.

For example, if your module requires `lib v1.2.0` and one of your dependencies requires `lib v1.4.0`, the build uses `v1.4.0`. Even if `v1.9.0` is available, it is not used until someone asks for it.

This makes builds **reproducible** without a separate lock file: the `go.mod` files themselves determine the result.

To see what was selected:

```bash
go list -m all                      # every module in the build, with versions
go mod why golang.org/x/text        # why a module is needed
go mod graph                        # the full requirement graph
```

### The major version rule

Starting at v2, the major version is **part of the module path**:

```go
import "github.com/some/lib/v2"
```

`github.com/some/lib` (v0 and v1) and `github.com/some/lib/v2` are, as far as Go is concerned, different modules. You can even use both in the same build during a migration. If you publish a v2 of your own module, the `module` line in its `go.mod` must end with `/v2`.

## Working on two modules at once

Sometimes you need to change a library and the app that uses it at the same time.

**Option 1: workspaces** (recommended for local development):

```bash
go work init ./app ./lib
```

This creates a `go.work` file that makes the `go` command use your local `lib` when building `app`. It is usually not committed.

**Option 2: `replace`** in `go.mod`:

```text
replace github.com/yourname/lib => ../lib
```

`replace` is simple, but it is easy to commit by accident and breaks builds for everyone else. Prefer workspaces.

## Common mistakes

### Forgetting `go mod tidy`

Code compiles on your machine, but CI fails with `missing go.sum entry` or `no required module provides package`. Run `go mod tidy` and commit both files.

### Expecting `replace` to apply to users of your library

`replace` and `exclude` only take effect in the **main module**, the one you run `go build` in. If you publish a library with a `replace` in its `go.mod`, the people who import it will not get your replacement.

### Publishing v2 without changing the module path

Tagging `v2.0.0` while `go.mod` still says `module github.com/you/lib` does not work as expected. The module path must become `github.com/you/lib/v2`, and importers update their imports.

### Choosing a module path that cannot be fetched

If your module path is `github.com/you/project` but the repository lives somewhere else, others cannot `go get` it. Match the module path to where the code is actually hosted.

## Key takeaways

- A module is a versioned tree of packages, defined by `go.mod`.
- `go.mod` lists minimum versions; `go.sum` locks hashes. Commit both, never hand-edit `go.sum`.
- `go mod tidy` is the command you run most.
- Minimal version selection makes builds reproducible without a lock file.
- Major versions 2 and above are part of the import path.

## Further reading

- [Go Modules Reference](https://go.dev/ref/mod), the complete specification
- [Tutorial: Create a Go module](https://go.dev/doc/tutorial/create-module)
- [Minimal version selection](https://research.swtch.com/vgo-mvs) by Russ Cox
- [Go modules: v2 and beyond](https://go.dev/blog/v2-go-modules) on the Go blog
- [Get familiar with workspaces](https://go.dev/blog/get-familiar-with-workspaces)
