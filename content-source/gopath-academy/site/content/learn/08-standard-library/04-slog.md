---
title: "8.4 log/slog structured logging"
linkTitle: "8.4 log/slog"
weight: 4
description: "Structured logging with log/slog: key-value attributes, text and JSON handlers, levels, With and groups, context-aware handlers, redaction with LogValuer, and MultiHandler in Go 1.26."
goVersion: "1.27"
---

## What you will learn

- Why structured logs beat formatted strings
- Logging with `slog.Info`, attributes and levels
- Text and JSON handlers, and `HandlerOptions`
- Adding common attributes with `With` and grouping them
- Changing the level at run time
- Redacting secrets with `LogValuer`
- Writing a handler that adds request IDs from the context
- Sending logs to several destinations with `NewMultiHandler` (Go 1.26+)

## The idea

`log.Printf("user %d bought %s for %d", id, item, price)` produces a string a human can read and a machine has to parse with regular expressions. **Structured logging** records each value as a separate key-value pair:

```text
level=INFO msg="purchase" user_id=7 item=book price=499
{"level":"INFO","msg":"purchase","user_id":7,"item":"book","price":499}
```

Log platforms (Loki, Elasticsearch, CloudWatch, Datadog) can then filter and aggregate on fields directly: every purchase over 400, every error for user 7. `log/slog` (Go 1.21) is the standard library's structured logger, and most new Go code uses it.

## Basics

{{< example file="examples/m08/slogging/slog_test.go" region="basics" run="go test -run Example_basics -v ./examples/m08/slogging" >}}

- Messages are short, constant descriptions of the event ("server started"). The variable parts go in **attributes**.
- Attributes are passed as alternating keys and values, or as typed `slog.Attr` values (`slog.String`, `slog.Int`, `slog.Duration`). Typed attrs avoid allocations and mismatched pairs; `go vet` checks the alternating form for odd counts.
- The four levels are `Debug`, `Info`, `Warn` and `Error`. Records below the handler's level are dropped cheaply.

The example helper removes the `time` attribute only to make the output stable for testing. Real logs keep it.

## Handlers

A `slog.Logger` formats nothing itself; it passes records to a `slog.Handler`. The standard library includes:

- `slog.NewTextHandler(w, opts)`: `key=value` pairs, good for local development.
- `slog.NewJSONHandler(w, opts)`: one JSON object per line, good for production log pipelines.
- `slog.DiscardHandler` (Go 1.24+): drops everything, useful in tests and libraries.

`slog.HandlerOptions` sets the minimum `Level`, `AddSource` (file and line), and `ReplaceAttr`, a hook to rename, reformat or remove attributes.

Set a default once in `main` with `slog.SetDefault(logger)`, and the package-level functions (`slog.Info(...)`) and the old `log` package both write through it.

## `With` and groups

{{< example file="examples/m08/slogging/slog_test.go" region="json" run="go test -run Example_jsonAndWith -v ./examples/m08/slogging" >}}

`logger.With(attrs...)` returns a child logger that adds those attributes to every record. Create one per request with the request ID, route and user, and pass it down (or put it in the context). Handlers pre-format `With` attributes once, so this is cheap.

`slog.Group` nests attributes under a key: `user.id` in text, `"user":{"id":7}` in JSON.

## Dynamic levels

{{< example file="examples/m08/slogging/slog_test.go" region="levels" run="go test -run Example_dynamicLevel -v ./examples/m08/slogging" >}}

A `slog.LevelVar` can be changed while the program runs, for example from an admin endpoint or a config reload, to turn on debug logging during an incident without a restart.

## Redacting secrets

{{< example file="examples/m08/slogging/slog_test.go" region="redact" run="go test -run Example_redaction -v ./examples/m08/slogging" >}}

A type that implements `slog.LogValuer` decides how it appears in logs, **everywhere** it is logged. Give secret-bearing types (tokens, passwords, API keys, full card numbers) a `LogValue` that returns a redacted value, and nobody can leak them into logs by accident. `ReplaceAttr` is a second line of defense for keys like `password`.

## Context and custom handlers

{{< example file="examples/m08/slogging/slog_test.go" region="context" run="go test -run Example_contextHandler -v ./examples/m08/slogging" >}}

Handlers compose like `io.Writer` wrappers. `ContextHandler` embeds another handler and adds the request ID from the context to every record. Calling `log.InfoContext(ctx, ...)` instead of `log.Info(...)` gives handlers access to the context, which is also how tracing libraries attach trace and span IDs (lesson 12.5).

## Several destinations (Go 1.26+)

{{< example file="examples/m08/slogging/multi_test.go" region="multi" run="go test -run Example_multiHandler -v ./examples/m08/slogging" >}}

`slog.NewMultiHandler` sends each record to every handler, for example human-readable text to the console and JSON to a file.

## What to log

- **Log events, not every function call.** Requests, state changes, errors with context, startup configuration (without secrets).
- **Log errors once**, at the edge where they are handled (lesson 5.6), with the attributes that identify the operation.
- **Keep keys consistent** across the codebase: `user_id` everywhere, not `userID` in one place and `uid` in another. A short list of shared key constants helps.
- **Do not log personal data** you do not need.

## Common mistakes

### Formatting values into the message

`slog.Info(fmt.Sprintf("user %d logged in", id))` throws away the structure. Use `slog.Info("user logged in", "user_id", id)`.

### Odd numbers of key-value arguments

`slog.Info("msg", "user_id")` produces a `!BADKEY` attribute. Run `go vet`, or use typed attrs.

### Expensive arguments at disabled levels

`log.Debug("state", "dump", expensiveDump())` computes the dump even when debug is off. Check `log.Enabled(ctx, slog.LevelDebug)` first, or make the value a `LogValuer` so it is only computed when logged.

### Logging secrets

Tokens in request headers, passwords in config structs. Use `LogValuer` and `ReplaceAttr`.

## Key takeaways

- Structured logs record key-value attributes that machines can query.
- `slog` loggers send records to handlers; use text locally and JSON in production.
- Add shared attributes with `With`, nest them with `Group`, and change levels at run time with `LevelVar`.
- Redact secrets with `LogValuer`; enrich records from the context with a wrapping handler.
- Go 1.26 adds `NewMultiHandler` for several destinations.

## Further reading

- [Package log/slog](https://pkg.go.dev/log/slog)
- [Structured Logging with slog](https://go.dev/blog/slog) on the Go blog
- [A Guide to Writing slog Handlers](https://github.com/golang/example/tree/master/slog-handler-guide)
