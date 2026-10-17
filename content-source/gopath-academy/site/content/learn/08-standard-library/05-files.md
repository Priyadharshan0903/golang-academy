---
title: "8.5 os, filepath, io/fs and embed"
linkTitle: "8.5 Files, fs and embed"
weight: 5
description: "Read and write files, write atomically, walk directory trees, program against the io/fs interface, embed files into binaries, and confine file access with os.Root."
goVersion: "1.27"
---

## What you will learn

- Reading, writing and appending files with `os`
- Building paths portably with `path/filepath`
- Writing files **atomically** so readers never see half-written data
- Walking directory trees with `filepath.WalkDir`
- The `io/fs` interface, and testing file code with `fstest.MapFS`
- Embedding files into your binary with `//go:embed`
- Restricting file access to one directory with `os.Root` (Go 1.24+)

## Reading and writing

{{< example file="examples/m08/files/files_test.go" region="readwrite" run="go test -run Example_readWrite -v ./examples/m08/files" >}}

- `os.ReadFile` and `os.WriteFile` handle open, read or write, and close in one call. They are the right choice for small files.
- `os.Open` (read only), `os.Create` (truncate or create) and `os.OpenFile` (any combination of flags) return an `*os.File`, which is an `io.Reader` and `io.Writer`, for streaming larger files. Close it with `defer f.Close()`, and check the error of `Close` for files you wrote (lesson 2.7).
- Permissions like `0o644` (owner read and write, others read) and `0o755` for directories are the usual values.
- Missing files return an error matching `fs.ErrNotExist`; check with `errors.Is`, never by message.

Use `filepath.Join`, `filepath.Dir`, `filepath.Base` and `filepath.Ext` for operating system paths, so the same code works with `\` on Windows. The plain `path` package is for slash-separated paths such as URLs and `io/fs` names.

## Atomic writes

`os.WriteFile` truncates the file, then writes. If the program crashes in between, or another process reads at the wrong moment, it sees an empty or partial file. For config files, state files and caches, write atomically:

{{< example file="examples/m08/files/files.go" region="atomic" >}}

{{< example file="examples/m08/files/files_test.go" region="atomicuse" run="go test -run Example_atomicWrite -v ./examples/m08/files" >}}

The temp file must be in the **same directory** (rename is only atomic within one filesystem), and `Sync` ensures the data is on disk before the rename makes it visible.

## Walking directories

{{< example file="examples/m08/files/files_test.go" region="walk" run="go test -run Example_walkDir -v ./examples/m08/files" >}}

`filepath.WalkDir` calls your function for every file and directory, in lexical order. Return `filepath.SkipDir` to skip a directory, or `filepath.SkipAll` to stop. Prefer it to the older `filepath.Walk`, which calls `os.Lstat` on every entry and is slower.

## `io/fs`: one interface for every filesystem

`fs.FS` is a read-only filesystem interface with a single method, `Open(name)`. Real directories (`os.DirFS`), embedded files, zip archives (`zip.Reader`) and test fixtures all implement it, and helpers like `fs.ReadFile`, `fs.Glob` and `fs.WalkDir` work on any of them:

{{< example file="examples/m08/files/files_test.go" region="fsabstraction" run="go test -run Example_fsInterface -v ./examples/m08/files" >}}

Write functions that read files against `fs.FS` instead of calling `os` directly. Tests can then pass an in-memory `fstest.MapFS`, with no temporary directories. Names in `io/fs` always use forward slashes and are relative (no leading `/`).

## Embedding files

{{< example file="examples/m08/files/files.go" region="embed" >}}

{{< example file="examples/m08/files/files_test.go" region="embeduse" run="go test -run Example_embed -v ./examples/m08/files" >}}

A `//go:embed` comment above a package-level variable copies files into the binary at build time:

- into a `string` or `[]byte` for a single file;
- into an `embed.FS` (which implements `fs.FS`) for patterns and whole directories.

This is how Go programs ship HTML templates, static web assets, SQL migrations and default config inside a single binary. Patterns are relative to the package directory, cannot reach outside the module, and skip files starting with `.` or `_` unless you use the `all:` prefix.

## Confining access with `os.Root`

Joining a user-supplied name to a base directory (`filepath.Join(uploads, name)`) is a classic **path traversal** hole: a name like `../../etc/passwd` escapes the directory. `os.Root` (Go 1.24) makes escaping impossible:

{{< example file="examples/m08/files/files_test.go" region="root" run="go test -run Example_osRoot -v ./examples/m08/files" >}}

Every operation on a `Root` is resolved inside its directory, including through symlinks. Go 1.25 added more methods (`MkdirAll`, `ReadFile`, `WriteFile`, `RemoveAll`, `Rename` and others). Use it whenever file names come from users, archives or network requests. `os.OpenInRoot(dir, name)` is a one-call shortcut.

## Common mistakes

### Building paths with string concatenation

`dir + "/" + name` breaks on Windows and makes traversal bugs easy. Use `filepath.Join`, and `os.Root` for untrusted names.

### Non-atomic writes of important files

A crash during `os.WriteFile` leaves a truncated file. Write to a temp file and rename.

### Forgetting that `embed` happens at build time

Changing an embedded file has no effect until you rebuild. And embedded files cannot be modified at run time.

### Not closing files

Each open file holds a file descriptor. Thousands of leaked ones produce `too many open files`.

## Key takeaways

- Use `os.ReadFile` and `os.WriteFile` for small files, `*os.File` streams for large ones.
- Build paths with `filepath`; write important files atomically via temp file, `Sync` and `Rename`.
- Walk trees with `filepath.WalkDir`.
- Program against `fs.FS` and test with `fstest.MapFS`.
- Embed assets with `//go:embed`; confine untrusted paths with `os.Root`.

## Further reading

- [Package os](https://pkg.go.dev/os), [path/filepath](https://pkg.go.dev/path/filepath) and [io/fs](https://pkg.go.dev/io/fs)
- [Package embed](https://pkg.go.dev/embed)
- [Traversal-resistant file APIs](https://go.dev/blog/osroot) on the Go blog
