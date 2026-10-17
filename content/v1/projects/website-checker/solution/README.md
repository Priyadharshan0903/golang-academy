# Sitecheck: a bounded website checker

A complete teaching project from Go by Doing. Standard library only; checked with Go 1.26.5.

## Run

```sh
go run ./cmd/sitecheck -workers 4 -timeout 2s -total 15s -file urls.example.txt
go run ./cmd/sitecheck -json https://example.com
printf 'https://example.com\n' | go run ./cmd/sitecheck -file -
```

Put flags before positional URL arguments. Output is sorted by URL; external HTTP results depend on the network. Exit 0 means every check returned a 2xx response, exit 1 means a failed check or interrupted run, and exit 2 means invalid usage. `go run` wraps a program's nonzero exit status; build a binary when scripting exact exit codes.

The total deadline covers the network checking stage after the URL input is read. Reading stdin or a local file is a separate stage and may block on its producer. Ctrl+C cancels the network run. Standard HTTP transports observe request contexts; an injected transport must also cooperate with cancellation for the checker to stop promptly.

## Verify

```sh
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/sitecheck
```

Tests inject HTTP transports, so they do not depend on the public internet. They exercise worker limits, cancellation, per-request deadlines, sorted results, invalid input, HTTP failures, response cleanup, and CLI input handling.

## Architecture

- `checker/checker.go`: validation, a fixed worker pool, request deadlines, body cleanup, sorted results.
- `checker/checker_test.go`: behavioral concurrency and HTTP tests.
- `cmd/sitecheck/main.go`: flags, input, cancellation, output, and exit codes.
- `cmd/sitecheck/main_test.go`: input/usage and reporting behavior.

The producer owns closing jobs; one coordinator closes results after all workers exit. The collector drains results even during cancellation. Each started request produces a result; URLs not started after cancellation may be absent. Individual failures live in Result, while run cancellation is returned as the separate error.

## Limits and extensions

- At most 8 workers and 1,000 URLs per run.
- Each body is discarded only up to 4 KiB; this bounds bytes, not wall-clock time. The request context covers body reading with the standard transport.
- This is a local CLI, not a public URL-fetching endpoint. Exposing it as a server would require a network-access policy, SSRF defenses, authentication/authorization, and rate limits.
- Redirects use the client's policy. A final status is reported after following allowed redirects.
- No automatic retries: decide retry safety and a total budget before adding them.
- Add per-host limits, richer error categories, durable schedules, or metrics as follow-up work.

Code: MIT. Teaching text: CC BY 4.0, attribution Go by Doing.
