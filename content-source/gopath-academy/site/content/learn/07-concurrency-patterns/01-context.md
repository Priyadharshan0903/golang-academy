---
title: "7.1 context.Context: cancellation, deadlines, values"
linkTitle: "7.1 context"
weight: 1
description: "How context carries cancellation and deadlines across function calls and goroutines, the rules for using it, cancellation causes, AfterFunc, WithoutCancel and request-scoped values."
goVersion: "1.27"
---

## What you will learn

- What problem `context.Context` solves
- `WithCancel`, `WithTimeout` and `WithDeadline`, and why you must always call `cancel`
- How cancellation flows from parent to child contexts
- Cancellation causes with `WithCancelCause` and `context.Cause`
- `AfterFunc` and `WithoutCancel`
- When to use context values, and when not to
- The conventions every Go API follows

## The idea

A request arrives at your server. To answer it, you query a database, call two other services and start a few goroutines. Then the client disconnects, or a deadline passes. All of that work should **stop**, now, everywhere.

`context.Context` is how Go propagates that "stop" signal (and deadlines, and a little request-scoped data) through every function and goroutine involved. It is an interface with four methods:

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{} // closed when the work should stop
	Err() error            // why: Canceled or DeadlineExceeded
	Value(key any) any
}
```

`Done()` is a channel that is **closed** on cancellation, the broadcast pattern from lesson 6.2. Any goroutine can `select` on it.

## Cancellation

{{< example file="examples/m07/contexts/contexts_test.go" region="cancel" run="go test -run Example_cancel -v ./examples/m07/contexts" >}}

- `context.Background()` is the root: never cancelled, no deadline, no values. Use it in `main`, in tests and at the top of background jobs.
- `context.WithCancel(parent)` returns a child context and a `cancel` function.
- Long-running work checks `ctx.Done()` in a `select`, and returns `ctx.Err()` when it is closed.

## Timeouts and deadlines

{{< example file="examples/m07/contexts/contexts_test.go" region="timeout" run="go test -run Example_timeout -v ./examples/m07/contexts" >}}

- `WithTimeout(parent, d)` cancels automatically after `d`; `WithDeadline(parent, t)` at time `t`.
- `ctx.Err()` is `context.DeadlineExceeded` for timeouts and `context.Canceled` for explicit cancels. Check with `errors.Is`.

**Always call `cancel`**, usually with `defer cancel()` right after creating the context, even when a timeout will fire anyway. It releases the context's resources as soon as the work is done instead of when the timer expires. `go vet` reports contexts whose `cancel` is never called.

## Cancellation flows down the tree

Contexts form a tree. Cancelling a parent cancels every descendant, but cancelling a child never affects its parent:

{{< example file="examples/m07/contexts/contexts_test.go" region="tree" run="go test -run Example_cancellationTree -v ./examples/m07/contexts" >}}

A child can only **shorten** its parent's deadline, never extend it. A function that adds `WithTimeout(ctx, time.Hour)` to a context that expires in one second still stops after one second.

## Why was it cancelled?

`ctx.Err()` only says "canceled" or "deadline exceeded". `WithCancelCause` (Go 1.20) lets you record the reason:

{{< example file="examples/m07/contexts/contexts_test.go" region="cause" run="go test -run Example_cancelCause -v ./examples/m07/contexts" >}}

`WithTimeoutCause` and `WithDeadlineCause` (Go 1.21) do the same for timeouts. Since Go 1.26, `signal.NotifyContext` also sets a cause identifying the signal that arrived. Causes make logs far more useful: "context canceled" becomes "server shutting down".

## `AfterFunc` and `WithoutCancel`

{{< example file="examples/m07/contexts/contexts_test.go" region="afterfunc" run="go test -run Example_afterFunc -v ./examples/m07/contexts" >}}

- `context.AfterFunc(ctx, f)` runs `f` in a new goroutine once `ctx` is done. Use it to tie cleanup to cancellation without writing a goroutine that waits on `Done()` yourself. It returns a `stop` function to unregister `f`.
- `context.WithoutCancel(ctx)` keeps the parent's values but is never cancelled. Use it sparingly, for work that must complete even if the request was abandoned, such as recording an audit event.

## Context values

`WithValue` attaches a key-value pair. Use it **only for request-scoped data that crosses API boundaries**: a request ID, a trace span, the authenticated user. Never for optional function parameters or dependencies like a database handle.

{{< example file="examples/m07/contexts/contexts_test.go" region="values" run="go test -run Example_values -v ./examples/m07/contexts" >}}

- Use an **unexported key type** so no other package can collide with your keys.
- Provide typed accessor functions (`WithRequestID`, `RequestID`) instead of making callers use `Value` and type assertions.
- Looking up a value walks up the tree, so deep chains of `WithValue` are slow. Store one struct rather than many separate values.

## The conventions

Every Go API that blocks or does I/O follows these rules. Follow them in yours:

1. **`ctx` is the first parameter**, named `ctx`: `func Fetch(ctx context.Context, url string) error`.
2. **Do not store a context in a struct.** Pass it to each call. (Exceptions are rare and documented.)
3. **Never pass a `nil` context.** Use `context.TODO()` when you have not wired one through yet; it is a searchable marker for "fix this later".
4. **The function that creates a context with `cancel` is responsible for calling it.**
5. **Respect it.** A function that takes a `ctx` and ignores it is lying to its callers.

The standard library honors context everywhere: `http.NewRequestWithContext`, `db.QueryContext`, `net.Dialer.DialContext`, `exec.CommandContext`. In an HTTP handler, `r.Context()` is cancelled when the client disconnects.

## Common mistakes

### Forgetting `defer cancel()`

Leaks a timer and a goroutine reference until the parent is cancelled, which for `Background` is never. `go vet`'s `lostcancel` check catches it.

### Checking `ctx` only at the start

A loop that checks `ctx.Err()` once and then runs for a minute cannot be cancelled. Check in every iteration, and pass `ctx` to every blocking call.

### Using `context.Background()` deep in the call stack

It cuts the cancellation chain: the work below it can no longer be stopped by the caller. Accept a `ctx` parameter and pass it down.

### Context values as a bag of parameters

Hidden dependencies that the compiler cannot check. If a function needs it, make it a parameter.

## Exercise

{{< exercise dir="m07/poll" >}}
Implement `WaitFor`, which polls a condition until it is true or the context ends, and `Retry`, which retries with exponential backoff but stops waiting the moment the context is cancelled. The tests measure timing, so a solution that sleeps without watching `ctx.Done()` will fail.
{{< /exercise >}}

## Key takeaways

- `context.Context` carries cancellation, deadlines and request-scoped values through calls and goroutines.
- `Done()` is closed on cancellation; select on it and return `ctx.Err()`.
- Create children with `WithCancel`, `WithTimeout` or `WithDeadline`, and always `defer cancel()`.
- Cancellation flows from parent to children; record reasons with `WithCancelCause` and read them with `context.Cause`.
- Pass `ctx` as the first parameter, never store it, never pass `nil`, and use values only for request-scoped data.

## Further reading

- [Package context](https://pkg.go.dev/context)
- [Go Concurrency Patterns: Context](https://go.dev/blog/context) on the Go blog
- [Contexts and structs](https://go.dev/blog/context-and-structs) on the Go blog
