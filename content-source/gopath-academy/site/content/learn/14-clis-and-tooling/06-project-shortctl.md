---
title: "14.6 Project: shortctl, a CLI for the shortener"
linkTitle: "14.6 Project: shortctl"
weight: 6
description: "Build a command-line client for the shortener API: layered configuration, text or JSON output, problem details turned into clear errors, proper exit codes, and tests that run the real API in-process."
goVersion: "1.27"
---

## The project

Services get command-line clients: `kubectl` for Kubernetes, `gh` for GitHub, `stripe` for Stripe. **shortctl** is the shortener's, and it uses every lesson of this module. It lives next to the server, in [`projects/shortener/cmd/shortctl`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/shortener/cmd/shortctl), so the two can share tests.

```text
$ shortctl create -code godoc https://go.dev/doc
http://localhost:8080/godoc

$ shortctl get godoc
http://localhost:8080/godoc -> https://go.dev/doc (created 2026-10-16 12:25 UTC)

$ shortctl -output json get godoc
{
  "code": "godoc",
  "url": "https://go.dev/doc",
  "short_url": "http://localhost:8080/godoc",
  "created_at": "2026-10-16T12:25:10.316877532Z"
}

$ shortctl create -code godoc https://go.dev; echo "exit $?"
shortctl: create "godoc": code already in use (HTTP 409)
exit 1

$ shortctl get; echo "exit $?"
shortctl: usage: get needs exactly one CODE
exit 2
```

That output is from a real run against the shortener service from Module 12.

## Try it

```bash
cd projects/shortener
go run ./cmd/shortener &                     # the API, in memory
go run ./cmd/shortctl create https://go.dev  # prints a short URL
go run ./cmd/shortctl -h
```

Point it at another server with `-server`, `SHORTCTL_SERVER`, or a config file at `<user config dir>/shortctl/config.json`:

```json
{"server": "https://sho.rt", "timeout": "5s"}
```

## The structure

`main` is three lines (lesson 14.1). `run` loads configuration, picks the command, and turns errors into exit codes in one place:

{{< example file="projects/shortener/cmd/shortctl/run.go" region="run" >}}

Configuration (`config.go`) is the layered loader of lesson 14.2: defaults, then the config file, then `SHORTCTL_*` variables, then only the flags actually given, found with `fs.Visit`.

Output has two modes. **Text** is a single line for people; **JSON** (`-output json`) is for scripts, so `shortctl -output json get godoc | jq -r .url` works. Offering both is cheap with a small `printer` type, and it saves users from parsing text with `awk`.

## The API client

{{< example file="projects/shortener/cmd/shortctl/client.go" region="client" >}}

Things to notice:

- **A timeout on the `http.Client`** (lesson 8.7), configurable, and the request carries the command's context, so Ctrl+C cancels a slow request.
- **Error responses are decoded.** The server answers errors with problem details (lesson 12.3), so the client shows the server's `detail`, `create "godoc": code already in use`, plus the status. If the body is not JSON (a proxy's HTML error page, say), it falls back to the status text.
- **The response body is limited** with `io.LimitReader`, so a misbehaving server cannot make the client read gigabytes.
- **Network errors say what was unreachable**: `cannot reach http://localhost:8080: ... connection refused`.
- **`url.PathEscape`** for the code in the path, even though the server validates codes: never build URLs from unescaped input.

## Testing against the real API

{{< example file="projects/shortener/cmd/shortctl/main_test.go" >}}

The tests start the **real** HTTP handler from `internal/httpapi` with an in-memory store on an `httptest.Server`, and run shortctl's `run` against it. This tests the client and the server's contract together: if the API changes a field name or a status code, these tests fail. It is only possible because both live in one module; for a client in another repository, the same idea becomes a contract test against a recorded or shared specification.

## Version

`shortctl version` prints `debug.ReadBuildInfo().Main.Version`. Built with `go install github.com/Priyadharshan0903/gopath-academy/projects/shortener/cmd/shortctl@v1.0.0`, it prints `v1.0.0`; built from a tagged git checkout with Go 1.24 or later, it prints the tag (lesson 14.4). No `-ldflags` needed.

## Your turn

1. Add `shortctl open CODE`, which opens the target URL in the browser (`xdg-open`, `open` or `rundll32` depending on `runtime.GOOS`).
2. Add `shortctl config`, printing the effective configuration and where each value came from (lesson 14.2's `Source` map).
3. Add a `-quiet` flag that prints only the short URL, for use in scripts: `url=$(shortctl -quiet create https://go.dev)`.
4. Add `shortctl` to the greet project's release setup (lesson 14.4): a second build in `.goreleaser.yaml`.

## Module 14 recap

You can now build command-line tools that behave the way users and scripts expect, with the standard `flag` package or Cobra; configure them from files, environment and flags in a predictable order; generate code with pinned tools or your own generator; and ship reproducible, versioned binaries for every platform with GoReleaser.

Next is **[Module 15: Go for cloud native](/learn/15-cloud-native/)**, where the shortener goes into a container and onto Kubernetes, and you write Go that talks to Kubernetes itself.
