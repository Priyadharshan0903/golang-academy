---
title: "5.3 Sentinel errors vs custom error types"
linkTitle: "5.3 Sentinels vs types"
weight: 3
description: "Three ways to let callers recognize errors: sentinel values, error types and behavior interfaces. When to use each, and what each commits you to."
goVersion: "1.27"
---

## What you will learn

- What **sentinel errors** are and when they fit
- When an **error type** is better than a sentinel
- How **behavior checks** keep errors opaque but still useful
- How a custom `Is` method makes one error match a category
- How each choice affects your package's API

## The idea

Once callers need to react differently to different failures, your package has to offer them a way to tell errors apart. There are three common designs, and each one becomes part of your public API, so choose deliberately.

## Sentinel errors

A sentinel is an exported variable holding a specific error value. Callers check for it with `errors.Is`:

{{< example file="examples/m05/kinds/kinds_test.go" region="sentinel" run="go test -run Example_sentinel -v ./examples/m05/kinds" >}}

Sentinels are named `ErrSomething`. The standard library is full of them: `io.EOF`, `fs.ErrNotExist`, `sql.ErrNoRows`, `context.Canceled`, `http.ErrServerClosed`.

**Use a sentinel when** the caller only needs to know **which** condition occurred, and there is no extra data to convey: not found, unauthorized, already exists, end of input.

## Error types

When callers need **data** about the failure, use a type:

{{< example file="examples/m05/kinds/kinds_test.go" region="typed" run="go test -run Example_errorType -v ./examples/m05/kinds" >}}

Examples from the standard library: `*fs.PathError` (operation and path), `*strconv.NumError` (function and input), `*net.OpError`, `*json.SyntaxError` (byte offset).

**Use an error type when** callers need fields: how long to wait, which field was invalid, which line failed, which HTTP status the upstream returned.

## Behavior checks: opaque errors

Sometimes you do not want callers to depend on either a specific value or a specific type. Instead, they ask **"can this error do X?"** through a small interface:

{{< example file="examples/m05/kinds/kinds_test.go" region="behavior" run="go test -run Example_behavior -v ./examples/m05/kinds" >}}

The concrete error type can stay unexported and change freely. `net.Error` uses this idea with its `Timeout() bool` method, so callers can retry timeouts without knowing which of many network error types they got.

**Use a behavior check when** many different errors share a property callers care about (retryable, timeout, permanent), especially across package boundaries.

## Matching a category with `Is`

An error type can define `Is(target error) bool` to say "I count as that error too":

{{< example file="examples/m05/kinds/kinds_test.go" region="is" run="go test -run Example_customIs -v ./examples/m05/kinds" >}}

This combines the two styles: callers who only care about the category check `errors.Is(err, ErrServer)`, while callers who need the status code use `errors.As` to get the `*HTTPError`.

## Choosing

| Caller needs to know | Use | Commits you to |
|---|---|---|
| Which condition happened | Sentinel `ErrX` | Keeping that variable and returning it (wrapped) in those cases |
| Details about the failure | Error type | Keeping the type and its fields |
| Whether a property holds | Behavior interface | Keeping the method |
| Nothing, just that it failed | Plain `fmt.Errorf` | Nothing |

Start with the last row. Many errors only need a good message. Add a sentinel or type when a real caller needs to branch on it.

## Common mistakes

### Too many sentinels

A package with thirty `ErrX` variables usually has callers that check only two. Every sentinel is API you must keep. Export only what callers need.

### Returning an error type by value in some places and by pointer in others

`errors.As` with `*MyError` will not find a `MyError` value. Be consistent; pointer receivers and pointer returns are the usual choice.

### Making sentinels mutable

`var ErrNotFound = errors.New(...)` is a variable, and someone could reassign it. That is a theoretical risk that the community accepts; the convention is simply never to assign to them.

### Leaking internal error types

Returning a database driver's error type straight to your callers makes them depend on your choice of driver. Translate it to your own sentinel or type at the package boundary.

## Key takeaways

- Sentinel errors (`ErrX`) identify a condition; check them with `errors.Is`.
- Error types carry data; extract them with `errors.As`.
- Behavior interfaces keep errors opaque but let callers test a property like "temporary".
- A custom `Is` method lets an error match a category.
- Everything you export becomes API. Start with plain errors and add structure when callers need it.

## Further reading

- [Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully) by Dave Cheney
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Package net: Error interface](https://pkg.go.dev/net#Error)
