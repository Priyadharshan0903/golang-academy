---
title: "5.5 panic and recover: when, and mostly when not"
linkTitle: "5.5 panic and recover"
weight: 5
description: "What a panic is, what causes them, how recover works, the narrow cases where panicking is right, and why a panic in any goroutine crashes the whole program."
goVersion: "1.27"
---

## What you will learn

- What happens when a program panics
- The common causes of run-time panics
- How `recover` works, and why it only works inside a deferred function
- The few situations where calling `panic` yourself is correct
- How to recover at boundaries, and why goroutines need special care

## The idea

A **panic** stops normal execution of the current goroutine. Deferred functions run as the stack unwinds, and if nothing recovers, the program **crashes** and prints the panic value and a stack trace for every goroutine.

Panics are for **bugs and impossible situations**, not for expected failures. A missing file, bad user input or a network timeout are expected; they return `error`. An index out of range, a nil pointer dereference or a broken invariant are bugs; they panic.

## Run-time panics

{{< example file="examples/m05/panics/panics_test.go" region="runtime" run="go test -run Example_runtimePanics -v ./examples/m05/panics" >}}

The runtime panics on: out-of-range indexes and slices, nil pointer dereferences, writes to nil maps, failed single-value type assertions, integer division by zero, closing a closed channel, sending on a closed channel, and calling `yield` after an iterator was told to stop.

## When it is right to panic

1. **Programmer errors that cannot be handled sensibly.** A function documented as "panics if n < 0" called with -1. The standard library does this, for example `strings.Repeat` with a negative count.
2. **Initialization of values fixed in the code.** `regexp.MustCompile`, `template.Must`: if the pattern written in the source is invalid, the program is broken and should fail immediately at startup.
3. **Truly impossible states.** A `default` branch of a switch over an enum that should never be reached, as an assertion.

{{< example file="examples/m05/panics/panics_test.go" region="must" run="go test -run Example_mustHelpers -v ./examples/m05/panics" >}}

The convention: functions that panic instead of returning an error are named `MustXxx`, and have an error-returning sibling `Xxx` for input that is not under your control.

## `recover`

`recover()` stops a panic and returns the value passed to `panic`. It only has an effect when called **directly inside a deferred function** during a panic. Anywhere else it returns `nil`.

{{< example file="examples/m05/panics/panics_test.go" region="boundary" run="go test -run Example_recoverAtBoundary -v ./examples/m05/panics" >}}

The deferred closure converts the panic into an error by setting the **named result** `err`, the technique from lesson 2.7.

### Recover at boundaries only

Recovering everywhere hides bugs. Recover at the **boundary** of an independent unit of work, where crashing everything would be worse than failing one unit:

- an HTTP request handler (the standard `net/http` server already recovers panics per request and logs them);
- a job in a worker pool;
- a call into a plugin or user-supplied callback.

After recovering, **log the panic with its stack trace** (`runtime/debug.Stack()`), fail that unit of work, and keep the service running.

### Recover selectively

{{< example file="examples/m05/panics/panics_test.go" region="repanic" run="go test -run Example_selectiveRecover -v ./examples/m05/panics" >}}

If you recover a panic you do not understand, **re-panic** it. Swallowing unknown panics leaves the program running in a state you cannot reason about.

## Panics and goroutines

`recover` only catches panics in the **same goroutine**. A panic in a goroutine you started, with no recover in that goroutine, **crashes the entire program**, no matter how many recovers exist in `main` or the HTTP handler that started it:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	go func() {
		var m map[string]int
		m["x"] = 1 // crashes the whole server, not just this request
	}()
}
```

If a goroutine runs code that might panic (third party callbacks, complex parsing), put a deferred recover at the top of that goroutine. Module 7 shows how worker pools and `errgroup` handle this.

## Common mistakes

### Using panic for ordinary errors

`panic(err)` on a failed database query turns a recoverable situation into a crash. Return the error.

### Using recover as try/catch

Wrapping ordinary code in recover to "handle" errors produces code that is slow, surprising and hard to follow. Go's error handling is `if err != nil`.

### Forgetting that `os.Exit` and `log.Fatal` skip defers and recovers

`log.Fatal` is not a panic. Deferred functions do not run. Use it only in `main`, if at all.

### Panicking across a package boundary

A library that panics on bad input forces every caller to add recovers. Libraries should return errors, and panic only on programmer errors that are documented.

## Key takeaways

- Panics are for bugs and impossible states; expected failures return `error`.
- `MustXxx` helpers panic on invalid constants at startup; use their error-returning siblings for runtime input.
- `recover` works only in a deferred function, in the same goroutine.
- Recover at boundaries (requests, jobs), log the stack, and re-panic what you do not understand.
- An unrecovered panic in any goroutine crashes the whole program.

## Further reading

- [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) on the Go blog
- [Handling panics](https://go.dev/ref/spec#Handling_panics) and [Run-time panics](https://go.dev/ref/spec#Run_time_panics) in the spec
- [Effective Go: Panic](https://go.dev/doc/effective_go#panic) and [Recover](https://go.dev/doc/effective_go#recover)
