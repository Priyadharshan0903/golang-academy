---
title: "1.6 The go command toolbox"
linkTitle: "1.6 The go command"
weight: 6
description: "The go subcommands you use every day, cross-compiling for other platforms, build constraints, and stamping version information into binaries."
goVersion: "1.27"
---

## What you will learn

- The `go` subcommands you will use daily, and what each is for
- How to cross-compile for another OS and CPU with two environment variables
- How build constraints select files per platform
- How to stamp a version into a binary with `-ldflags`, and how to read build info back

## The idea

Go ships one tool, `go`, that builds, tests, formats, documents, downloads dependencies and more. There is no separate build system to learn for most projects. Run `go help` to see everything; this lesson covers what matters in practice.

## Everyday commands

| Command | What it does |
|---|---|
| `go run .` | Compile and run the package in the current directory. |
| `go build ./...` | Compile every package in the module. Libraries are compiled and discarded; this is a quick "does it all build?" check. |
| `go build -o bin/app ./cmd/app` | Build one binary to a chosen path. |
| `go test ./...` | Run all tests. Covered fully in Module 9. |
| `go install ./cmd/app` | Build and copy the binary to `$GOBIN` (or `$GOPATH/bin`). |
| `go install example.com/tool@v1.4.0` | Install a tool at a specific version, outside of any project. |
| `go vet ./...` | Report suspicious code. See lesson 1.7. |
| `go fmt ./...` | Format code (runs `gofmt -l -w` on the packages). |
| `go doc strings.TrimSpace` | Show documentation in the terminal. |
| `go env` | Show or change Go settings. |
| `go list ./...` | List packages. With `-json` or `-f`, a powerful query tool for scripts. |
| `go mod tidy` | Sync `go.mod` and `go.sum` with your imports (lesson 1.4). |
| `go clean -cache -testcache` | Clear the build and test caches when something seems stale. |

The pattern `./...` means "this directory and everything below it". It shows up everywhere.

Go caches build results aggressively. A second `go build` or `go test` with no changes is almost instant, and `go test` prints `(cached)` when it reuses a previous result.

## Useful build flags

| Flag | Use |
|---|---|
| `-o path` | Output file or directory. |
| `-v` | Print package names as they are compiled. |
| `-race` | Enable the race detector (Module 6). |
| `-trimpath` | Remove local file system paths from the binary. Use it for release builds. |
| `-ldflags "..."` | Pass flags to the linker; used for version stamping below. |
| `-tags a,b` | Enable build tags (see build constraints). |

## Cross-compiling

To build for another platform, set `GOOS` (operating system) and `GOARCH` (CPU architecture):

{{< example file="examples/m01/platform/main.go" run="go run ./examples/m01/platform" >}}

```bash
GOOS=linux   GOARCH=amd64 go build -o dist/platform-linux-amd64 ./examples/m01/platform
GOOS=linux   GOARCH=arm64 go build -o dist/platform-linux-arm64 ./examples/m01/platform
GOOS=darwin  GOARCH=arm64 go build -o dist/platform-darwin-arm64 ./examples/m01/platform
GOOS=windows GOARCH=amd64 go build -o dist/platform-windows-amd64.exe ./examples/m01/platform
```

No extra toolchains, no Docker. Run `go tool dist list` to see every supported combination.

### CGO and static binaries

Some packages can use C code through **cgo**. When cgo is involved, cross-compiling needs a C cross-compiler, and binaries may depend on the system C library. For most services and CLIs you want a pure Go, fully static binary:

```bash
CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app
```

A static binary like this runs in a minimal container image with nothing else inside. We use exactly this in Module 15 and for the server that hosts this site.

## Build constraints: per-platform files

Sometimes code must differ per platform. Go selects files at build time in two ways.

**File name suffixes**: `paths_windows.go` is only compiled for Windows, `paths_linux.go` only for Linux, `cpu_arm64.go` only for arm64. The pattern is `name_GOOS.go`, `name_GOARCH.go` or `name_GOOS_GOARCH.go`.

**`//go:build` lines** at the top of a file, before `package`:

```go
//go:build linux || darwin

package paths
```

Constraints support `&&`, `||`, `!` and parentheses, and can use custom tags enabled with `-tags`. This course uses that for exercises: solutions start with `//go:build solution`, so they are only compiled when you run `go test -tags solution`.

## Stamping a version into the binary

A binary should be able to tell you what it is. The linker flag `-X` sets the value of a **package-level string variable** at build time:

{{< example file="examples/m01/buildinfo/main.go" region="ldflags" >}}

```bash
go build -ldflags "-X main.version=1.2.0 -X main.commit=$(git rev-parse --short HEAD)" \
  -o buildinfo ./examples/m01/buildinfo
./buildinfo
```

```text
version=1.2.0 commit=3f9c2ab
module version: (devel)
```

The second line depends on how the binary was built: `(devel)` outside a git checkout, a version or pseudo-version derived from git inside one, or the exact tag after `go install module@version`.

Without `-ldflags`, the defaults from the source are used. The test proves it:

{{< example file="examples/m01/buildinfo/main_test.go" >}}

For variables outside `package main`, use the full import path: `-X github.com/you/app/internal/version.Version=1.2.0`. The checkpoint project does this.

Two more linker flags you will often see in release builds: `-s -w` strip the symbol table and debug information to make the binary smaller (at the cost of less useful crash dumps and debugger support).

### Build info the toolchain records for free

Go also embeds module and build metadata in every binary. You can read it from inside the program:

{{< example file="examples/m01/buildinfo/main.go" region="buildinfo" >}}

And from outside, for any Go binary on your system:

```bash
go version -m ./buildinfo
```

This prints the Go version used, the module path and version, every dependency with its version and hash, and build settings like `GOOS`, `GOARCH`, `CGO_ENABLED` and the VCS commit. It is a great way to audit what is inside a binary you did not build.

## Common mistakes

### Using `-X` on a constant

`-X` only changes string **variables**. `const version = "dev"` silently stays `"dev"`. Use `var`.

### A typo in the `-X` path

If the import path or variable name in `-X pkg.name=value` is wrong, the linker does not complain; the value just does not change. Check by running `./app -version` after every change to your build script.

### Forgetting `CGO_ENABLED=0` for containers

A binary built with cgo on a glibc system may fail to start in Alpine or distroless images with a confusing "not found" error. For containers, build with `CGO_ENABLED=0` unless you know you need cgo.

### Committing binaries

Add `bin/` and `dist/` to `.gitignore`. Binaries belong in releases, not in the repository.

## Key takeaways

- One `go` command covers building, testing, formatting, docs and dependencies.
- `./...` means "every package from here down".
- Cross-compiling is just `GOOS` and `GOARCH`. Add `CGO_ENABLED=0` for static binaries.
- Build constraints (`_linux.go` suffixes and `//go:build` lines) select files per platform or tag.
- `-ldflags "-X pkg.var=value"` stamps versions; `go version -m` reads build info from any Go binary.

## Further reading

- [Command go](https://pkg.go.dev/cmd/go), full reference
- [Build constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints)
- [Command link](https://pkg.go.dev/cmd/link), including `-X`
- [runtime/debug.ReadBuildInfo](https://pkg.go.dev/runtime/debug#ReadBuildInfo)
