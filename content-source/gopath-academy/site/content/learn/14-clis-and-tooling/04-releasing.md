---
title: "14.4 Releasing binaries"
linkTitle: "14.4 Releasing binaries"
weight: 4
description: "Ship Go binaries people can trust: the version and VCS information the toolchain embeds (Go 1.24 stamps the module version from git tags), cross-compilation, reproducible builds verified by checksum, GoReleaser, and checking for known vulnerabilities."
goVersion: "1.27"
---

## What you will learn

- What the Go toolchain records in every binary, and how to read it with `go version -m`
- How Go 1.24 stamps the main module's version from your git tag
- When you still need `-ldflags -X`
- Cross-compiling for every platform from one machine
- Reproducible builds, checked by comparing checksums
- Releasing with GoReleaser, and scanning for vulnerabilities with `govulncheck`

## What a binary knows about itself

Every Go binary carries its build information: the Go version, the module path and version, every dependency with its version and checksum, the build settings, and version control details. `go version -m` prints it. Here is the greet CLI (lesson 1.8), built from a git checkout tagged `v1.4.0`:

```text
$ go version -m dist/greet_linux_amd64_v1/greet
dist/greet_linux_amd64_v1/greet: go1.27.2
	path	github.com/Priyadharshan0903/gopath-academy/projects/greet/cmd/greet
	mod	github.com/Priyadharshan0903/gopath-academy	v1.4.0
	build	-buildmode=exe
	build	-compiler=gc
	build	-trimpath=true
	build	DefaultGODEBUG=containermaxprocs=0,...,updatemaxprocs=0,...
	build	CGO_ENABLED=0
	build	GOARCH=amd64
	build	GOOS=linux
	build	GOAMD64=v1
	build	vcs=git
	build	vcs.revision=0c493c488ab4c76d293de9b7580151f174c04a6b
	build	vcs.time=2026-10-16T12:19:20Z
	build	vcs.modified=false
```

A program reads the same data with `runtime/debug.ReadBuildInfo()` (lesson 1.8 showed it). Three lines deserve attention:

- **`mod ... v1.4.0`.** Since **Go 1.24**, `go build` sets the main module's version from the VCS tag on the current commit, or a pseudo-version for untagged commits, with `+dirty` added when the working tree has uncommitted changes. Before Go 1.24, local builds always said `(devel)`. A test build of the same tree with an untracked `dist/` folder in it said `v1.4.0+dirty`, which is why build output directories belong in `.gitignore`.
- **`vcs.revision`, `vcs.time`, `vcs.modified`** are recorded automatically too (since Go 1.18), unless you build with `-buildvcs=false`.
- **`DefaultGODEBUG`** lists the compatibility settings this binary uses. The course's main module says `go 1.24` in `go.mod`, so even though Go 1.27 built it, the binary keeps Go 1.24's behavior, including `containermaxprocs=0` and `updatemaxprocs=0`: the container-aware `GOMAXPROCS` of lesson 11.1 is **off**. Raising the `go` line in `go.mod` turns it on. This line is how you check.

So for many tools, `debug.ReadBuildInfo()` already gives a correct version, with nothing to configure. The shortctl client (lesson 14.6) does exactly that.

## When to use -ldflags -X

`-ldflags "-X importpath.name=value"` sets a package-level string variable at link time. It is still useful when:

- you want a version string formatted your own way, or a build date (build info records the commit time, not the build time);
- the build happens outside a git checkout, for example from a source tarball, where there is no VCS information;
- you build with a release tool that knows the version better than the checkout does.

The greet project uses both: `-X` for a human-friendly version line, and build info is there anyway.

## Cross-compiling

The Go toolchain compiles for any supported platform from any other, with two environment variables:

```bash
GOOS=windows GOARCH=amd64 go build -o greet.exe ./cmd/greet
GOOS=darwin  GOARCH=arm64 go build -o greet-mac ./cmd/greet
go tool dist list    # every supported GOOS/GOARCH pair
```

The one thing that breaks this is **cgo**: C code needs a C cross-compiler for the target. With `CGO_ENABLED=0`, Go uses pure Go implementations (for example of DNS resolution) and produces a **static binary** with no dependency on the system's C library, which also runs in the smallest container images (lesson 15.1). Prefer pure Go dependencies for tools you distribute; SQLite drivers, for instance, exist in both forms.

## Reproducible builds

Go builds are reproducible: the same source, Go version and settings produce byte-identical binaries on any machine (and since Go 1.21, so are the Go toolchain downloads themselves). Two things you add:

- **`-trimpath`** removes your local file paths from the binary (they otherwise appear in stack traces and break reproducibility between machines).
- **Fixed timestamps** in archives. GoReleaser's `mod_timestamp: "{{ .CommitTimestamp }}"` sets file times to the commit time instead of "now".

The proof is cheap. Building all six greet archives twice, with the build cache wiped in between, gave identical checksums:

```text
$ diff first/checksums.txt dist/checksums.txt && echo identical
identical
```

Reproducibility lets anyone verify that a published binary was built from the source it claims, which is increasingly expected for anything security sensitive.

## GoReleaser

Building six platforms, archiving, writing checksums, generating a changelog and publishing a GitHub release by hand is tedious. **GoReleaser** does all of it from one config file:

{{< example file="projects/greet/.goreleaser.yaml" >}}

Try it without publishing anything:

```bash
cd projects/greet
goreleaser check                     # validate the config
goreleaser release --snapshot --clean
```

```text
  • building binaries
    • building    paths=cmd/greet binaries=greet target=linux_amd64_v1
    ...
  • archives
    • archiving   name=dist/greet_1.4.0_linux_amd64.tar.gz
    ...
  • calculating checksums
  • release succeeded
```

In a repository of its own, the release flow is: push a tag (`git tag v1.4.0 && git push --tags`), and a CI job runs `goreleaser release`, which builds everything and creates the GitHub release with the archives and `checksums.txt` attached. This course's repository holds many projects, so its workflow (`.github/workflows/greet-release.yml`) runs in snapshot mode and uploads the archives as a build artifact instead; the comments in the file show the two lines to change for real releases. GoReleaser can also publish Homebrew formulas, Linux packages, container images and signed checksums.

## Known vulnerabilities

Before releasing, check your dependencies and the standard library for known vulnerabilities:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

`govulncheck` uses the Go vulnerability database and, unlike most scanners, analyzes **which functions you actually call**: a vulnerable package you import but whose vulnerable function you never reach is reported separately, as informational. That keeps the signal high. The course's CI runs it on every module.

Keep the Go toolchain itself up to date too: security fixes for the standard library ship in minor releases (1.27.1, 1.27.2, ...), and a rebuilt binary is the only way to get them.

## Common mistakes

### Version "dev" in released binaries

Release from a clean, tagged checkout, or stamp the version explicitly.

### Build directories inside the repository, not ignored

They make every build `+dirty`. Add them to `.gitignore`.

### cgo enabled by accident

A dependency that needs cgo silently breaks cross-compilation and static binaries. Build releases with `CGO_ENABLED=0` and notice when that fails.

### No checksums

Users cannot verify downloads. Always publish `checksums.txt`.

## Key takeaways

- `go version -m` (and `debug.ReadBuildInfo`) show a binary's Go version, dependencies, build settings and VCS revision.
- Since Go 1.24, the module version comes from your git tag automatically, with `+dirty` for uncommitted changes.
- `DefaultGODEBUG` reveals which compatibility behaviors your `go.mod` line selected.
- Cross-compile with `GOOS`/`GOARCH`; `CGO_ENABLED=0` gives static, portable binaries.
- `-trimpath` plus fixed timestamps give reproducible builds; verify by comparing checksums.
- GoReleaser turns a tag into a full release; `govulncheck` checks what you actually call.

## Further reading

- [Go 1.24 release notes: go build version stamping](https://go.dev/doc/go1.24#go-command)
- [Perfectly Reproducible, Verified Go Toolchains](https://go.dev/blog/rebuild) on the Go blog
- [GoReleaser documentation](https://goreleaser.com/)
- [govulncheck tutorial](https://go.dev/doc/tutorial/govulncheck)
