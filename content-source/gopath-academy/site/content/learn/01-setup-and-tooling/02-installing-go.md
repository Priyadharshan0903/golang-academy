---
title: "1.2 Installing Go and setting up your editor"
linkTitle: "1.2 Installing Go"
weight: 2
description: "Install the Go toolchain, understand GOPATH and GOROOT, set up gopls in your editor, and learn how Go manages toolchain versions."
goVersion: "1.27"
---

## What you will learn

- How to install Go on macOS, Linux and Windows
- What `GOROOT`, `GOPATH`, `GOBIN` and the module cache are
- How to set up your editor with `gopls`, the official language server
- How the `go` command can switch toolchain versions for you

## Install Go

Always download from the official page: [go.dev/dl](https://go.dev/dl/). It lists the current stable release for every platform.

{{< tabs >}}
{{< tab name="macOS" >}}
Download the `.pkg` installer from go.dev/dl and run it. It installs Go into `/usr/local/go` and adds `/usr/local/go/bin` to your `PATH`.

If you use Homebrew, this also works:

```bash
brew install go
```
{{< /tab >}}
{{< tab name="Linux" >}}
Download the `linux-amd64` (or `linux-arm64`) archive from go.dev/dl, then replace any existing install and extract it:

```bash
# Replace go1.27.x with the exact file name you downloaded.
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.27.x.linux-amd64.tar.gz
```

Add Go to your `PATH` in `~/.profile`, `~/.bashrc` or `~/.zshrc`:

```bash
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
```

Open a new terminal so the change takes effect.

{{< callout type="warning" >}}
Always remove the old `/usr/local/go` before extracting a new version. Extracting on top of an old install mixes files from two versions and causes confusing errors.
{{< /callout >}}
{{< /tab >}}
{{< tab name="Windows" >}}
Download the `.msi` installer from go.dev/dl and run it. It installs Go and updates your `PATH`. Restart any open terminals afterwards.

With winget:

```powershell
winget install GoLang.Go
```
{{< /tab >}}
{{< /tabs >}}

Check that it worked:

```bash
go version
```

```text
go version go1.27.x linux/amd64
```

The exact patch version and platform will match your machine.

## Where Go keeps things

Run `go env` to see every setting the `go` command uses. These are the ones worth knowing on day one:

```bash
go env GOROOT GOPATH GOBIN GOMODCACHE
```

| Variable | What it is | Typical value |
|---|---|---|
| `GOROOT` | Where the Go toolchain itself is installed. You almost never set this. | `/usr/local/go` |
| `GOPATH` | Your Go workspace. Today it mainly holds downloaded modules and installed binaries. | `$HOME/go` |
| `GOBIN` | Where `go install` puts binaries. Empty means `$GOPATH/bin`. | empty |
| `GOMODCACHE` | Downloaded module source code, shared by all your projects. | `$HOME/go/pkg/mod` |

Add `$HOME/go/bin` to your `PATH` (the Linux snippet above already does). Otherwise tools you install with `go install` will not be found.

{{< callout type="info" >}}
**A bit of history.** Before Go 1.11, all Go code had to live inside `GOPATH/src`, and older tutorials still describe that layout. With modules (lesson 1.4) your projects can live anywhere on disk. You can ignore advice that tells you to put code in `GOPATH`.
{{< /callout >}}

You can change settings permanently with `go env -w`, which writes to a per-user config file instead of your shell profile:

```bash
go env -w GOBIN=$HOME/bin
go env -u GOBIN   # undo, back to the default
```

## Set up your editor

All good Go editor support is built on **gopls** (pronounced "go please"), the official language server. It provides completion, go to definition, find references, rename, inline errors and formatting on save.

{{< tabs >}}
{{< tab name="VS Code" >}}
Install the official **Go** extension (published by the Go team at Google). On first use it offers to install `gopls` and other tools. Accept.

Recommended settings:

```json
{
  "editor.formatOnSave": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```
{{< /tab >}}
{{< tab name="GoLand" >}}
JetBrains GoLand works out of the box with its own analysis engine. Enable **Settings → Tools → Actions on Save → Reformat code** and **Optimize imports**.
{{< /tab >}}
{{< tab name="Neovim / others" >}}
Install `gopls` yourself and point your LSP client at it:

```bash
go install golang.org/x/tools/gopls@latest
```

Any editor with LSP support (Neovim, Helix, Zed, Emacs, Sublime) can use it.
{{< /tab >}}
{{< /tabs >}}

Whatever editor you choose, make sure **format on save** is on. In Go, formatting is not a matter of taste; lesson 1.7 explains why.

## Toolchains: let `go` manage versions

Since Go 1.21, the `go` command can download and run a different Go version for you. A project's `go.mod` file says which version of Go it needs:

```text
go 1.27
```

If your installed Go is older than that, the default setting (`GOTOOLCHAIN=auto`) makes the `go` command download the required toolchain into the module cache and use it automatically. This means:

- You rarely need a version manager like `gvm` or `asdf` for Go.
- A team can upgrade Go by editing one line in `go.mod`.

You can see which toolchain is actually being used:

```bash
go version          # inside a project, shows the toolchain selected for it
go env GOTOOLCHAIN  # usually "auto"
```

To try a specific release side by side without changing your main install:

```bash
go install golang.org/dl/go1.26.0@latest
go1.26.0 download
go1.26.0 version
```

## Verify your setup

Clone this course and run the examples. If this passes, your setup is ready for the rest of the module.

```bash
git clone https://github.com/Priyadharshan0903/gopath-academy.git
cd gopath-academy
go test ./examples/...
```

```text
ok  	github.com/Priyadharshan0903/gopath-academy/examples/m01/buildinfo	0.004s
ok  	github.com/Priyadharshan0903/gopath-academy/examples/m01/greeting	0.003s
...
```

## Common mistakes

### `go: command not found` after installing

The installer or archive put Go on disk, but your shell's `PATH` does not include it yet. Open a new terminal. On Linux, check that the `export PATH=...` line is in the file your shell actually reads.

### Installed tools are "not found"

You ran `go install some/tool@latest` successfully, but typing `tool` fails. The binary is in `$(go env GOPATH)/bin` (or `GOBIN`), which is not on your `PATH`.

### Two Go versions fighting

A Homebrew install and a manual `/usr/local/go` install on the same Mac, or an old distro package on Linux, can shadow each other. Run `which -a go` to see every `go` on your `PATH` and remove the one you do not want.

### Setting `GOROOT` by hand

Setting `GOROOT` manually is almost never necessary and often breaks things after an upgrade. If you set it in your shell profile, remove it.

## Key takeaways

- Install Go from go.dev/dl and add both the Go `bin` folder and `$HOME/go/bin` to your `PATH`.
- `GOPATH` now mostly holds the module cache and installed tools. Your projects can live anywhere.
- `gopls` powers editor support everywhere. Turn on format on save.
- The `go` line in `go.mod` plus `GOTOOLCHAIN=auto` lets the `go` command fetch the right toolchain for each project.

## Further reading

- [Download and install](https://go.dev/doc/install) on go.dev
- [Managing Go installations](https://go.dev/doc/manage-install)
- [Go toolchains](https://go.dev/doc/toolchain), how `GOTOOLCHAIN` works
- [gopls documentation](https://go.dev/gopls/)
