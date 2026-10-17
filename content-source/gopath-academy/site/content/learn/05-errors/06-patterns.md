---
title: "5.6 Error handling patterns in real codebases"
linkTitle: "5.6 Error patterns"
weight: 6
description: "Where errors should be handled, logged, translated and retried in a real service, plus patterns that reduce repetitive error checks."
goVersion: "1.27"
---

## What you will learn

- Where in a program each error should be handled
- The "handle once" rule: log **or** return, not both
- How to translate domain errors into HTTP status codes at the edge
- How to retry only the errors that retrying can fix
- Patterns that keep repetitive error checks readable

## The idea

Errors flow **up** the call stack. Each layer has a different job:

| Layer | Job with errors |
|---|---|
| Low level (database, file, HTTP client) | Return errors with context, using sentinels or types where callers need them |
| Domain (business logic) | Translate low level errors into domain errors (`ErrNotFound`, `ErrConflict`), add context |
| Edge (HTTP handler, CLI `main`, worker loop) | **Decide**: map to a status code or exit code, log once, show a safe message to the user |

Most code is in the first two layers and simply returns. Only the edge handles.

## Handle each error once

```go
// Wrong: the error is logged here AND by the caller.
if err != nil {
	log.Printf("could not load user: %v", err)
	return err
}
```

Logging and returning produces duplicate log lines, often with less context than the final one. The rule:

- If you **return** the error, add context and do not log.
- If you **handle** it (log it, fall back, show it to the user), do not return it.

The edge of the program logs, because that is where you know the full context (request ID, user, route).

## Translate at the edge

{{< example file="examples/m05/patterns/patterns_test.go" region="mapping" run="go test -run Example_mapAtTheEdge -v ./examples/m05/patterns" >}}

One function maps domain errors to HTTP status codes. Handlers call it, the domain layer never imports `net/http`, and the mapping lives in exactly one place. Anything unrecognized is a 500, and the details stay in your logs rather than in the response. The URL shortener's HTTP layer in Module 12 uses exactly this.

## Retry only what can succeed

{{< example file="examples/m05/patterns/patterns_test.go" region="retry" run="go test -run Example_retry -v ./examples/m05/patterns" >}}

Retrying a validation error wastes time and load: it will fail the same way every time. Classify errors (sentinel, type or behavior check from lesson 5.3) and retry only transient ones, with **exponential backoff**. Real retry loops also add jitter and respect `context` cancellation, which Module 7 adds.

## Reducing repetition

Go code has many `if err != nil` checks, and that is fine: they are explicit and easy to scan. When the same check repeats many times in a row, two patterns help.

### Remember the first error

{{< example file="examples/m05/patterns/patterns_test.go" region="errwriter" run="go test -run Example_errWriter -v ./examples/m05/patterns" >}}

This is the idea behind `bufio.Writer` and `bufio.Scanner`: operations become no-ops after the first failure, and you check once at the end (`w.Flush()`, `sc.Err()`).

### Wrap once with a deferred helper

```go
func (s *Service) Sync(id string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("sync %s: %w", id, err)
		}
	}()
	// many return statements, each returning a bare err
}
```

Every error that leaves `Sync` gets the same context, added in one place (lesson 2.7).

## A checklist for reviewing error handling

1. Is every returned error checked?
2. Does each returned error add context the caller does not have?
3. Is each error logged at most once?
4. Are callers checking causes with `errors.Is` or `errors.As`, never by string?
5. Are internal details (SQL, stack traces, secrets) kept out of user-facing messages?
6. Are only transient errors retried, with backoff and a limit?
7. Do goroutines report errors back instead of dropping them? (Module 7)

## Common mistakes

### Logging deep in the stack

Library code that logs errors forces its logging style on every user and duplicates messages. Return errors; let the application log.

### Exposing internal errors to users

A 500 response body containing `pq: duplicate key value violates unique constraint "users_email_key"` leaks your schema. Map it to a domain error and a safe message.

### Retrying forever

A retry loop without a limit or a deadline turns an outage into a pile-up of stuck requests.

## Key takeaways

- Low layers return errors with context; domain layers translate them; the edge decides and logs.
- Handle each error once: log it or return it, never both.
- Map domain errors to status codes in one place at the edge.
- Retry only transient errors, with backoff and limits.
- Sticky-error writers and deferred wrapping reduce repetition without hiding errors.

## Further reading

- [Errors are values](https://go.dev/blog/errors-are-values) by Rob Pike
- [Error handling and Go](https://go.dev/blog/error-handling-and-go)
- [Google Go Style Guide: Errors](https://google.github.io/styleguide/go/best-practices#error-handling)
