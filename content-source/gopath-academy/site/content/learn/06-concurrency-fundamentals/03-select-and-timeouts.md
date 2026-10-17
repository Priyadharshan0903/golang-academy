---
title: "6.3 select, timeouts and time.After pitfalls"
linkTitle: "6.3 select and timeouts"
weight: 3
description: "Wait on several channel operations with select, implement timeouts and non-blocking sends, use timers and tickers correctly, and test time-dependent code with testing/synctest."
goVersion: "1.27"
---

## What you will learn

- How `select` waits on several channel operations at once
- How to add a timeout to any blocking operation
- Non-blocking sends and receives with `default`
- How a nil channel disables a `select` case
- Timers and tickers, and what changed about them in Go 1.23
- Testing time-based code deterministically with `testing/synctest` (Go 1.25+)

## The idea

`select` is like a `switch` for channel operations. It blocks until **one** of its cases can proceed, then runs that case. If several are ready at the same time, it picks one **at random**, so no case can starve the others.

{{< example file="examples/m06/selects/selects_test.go" region="select" run="go test -run Example_select -v ./examples/m06/selects" >}}

## Timeouts

Combining a real operation with a timer channel gives you a timeout:

{{< example file="examples/m06/selects/selects_test.go" region="timeout" run="go test -run Example_timeout -v ./examples/m06/selects" >}}

`time.After(d)` returns a channel that receives a value after `d`. Notice that `slowLookup` uses a **buffered** channel of size one. When the timeout wins, nobody ever reads the result; with an unbuffered channel the goroutine would block on its send **forever**, leaking it. A buffer of one lets it deliver and exit. This small detail is one of the most common sources of goroutine leaks (lesson 7.6).

In real code, timeouts are usually expressed with `context.WithTimeout` instead, which composes across function calls and goroutines (lesson 7.1). The `select` mechanics are the same: you wait on `ctx.Done()`.

## Non-blocking operations

A `default` case runs immediately if no other case is ready:

{{< example file="examples/m06/selects/selects_test.go" region="default" run="go test -run Example_nonBlocking -v ./examples/m06/selects" >}}

This is useful for dropping work when a queue is full (load shedding), for polling, and for "try" operations. Do not put a `default` in a `select` inside a busy loop: it turns blocking into spinning at 100% CPU.

## Disabling a case with a nil channel

Operations on a nil channel block forever, so a nil channel's case in a `select` can never be chosen. Setting a channel variable to `nil` is the idiomatic way to stop listening to it:

{{< example file="examples/m06/selects/selects_test.go" region="nilcase" run="go test -run Example_nilChannelDisablesCase -v ./examples/m06/selects" >}}

When one input closes, its variable is set to `nil`, and the loop keeps receiving from the other until both are done.

## Timers and tickers

{{< example file="examples/m06/selects/selects_test.go" region="ticker" run="go test -run Example_ticker -v ./examples/m06/selects" >}}

- `time.NewTimer(d)` fires once; `time.NewTicker(d)` fires repeatedly. Both deliver on their channel `C`.
- Call `Stop` when you are done, typically with `defer`.
- `time.Tick` is a convenience for long-lived tickers you never stop.

### What changed in Go 1.23

Older articles warn that `time.After` inside a loop leaks memory until each timer fires, and describe complicated rules for `Reset` and `Stop`. Go 1.23 changed timers so that:

- timers and tickers that are no longer referenced are garbage collected **even if not stopped**;
- timer channels are **synchronous** (unbuffered), so after `Stop` or `Reset` returns, no stale value will be received.

Go 1.27 removed the `asynctimerchan` setting that could bring back the old behavior. Calling `Stop` is still good practice (it frees resources sooner and documents intent), and `time.After` in a hot loop still allocates a timer per iteration, so prefer one `time.NewTimer` or a `context` deadline in loops.

## Testing time with `testing/synctest`

Code with timeouts is hard to test: real sleeps make tests slow, and short sleeps make them flaky. Go 1.25 added `testing/synctest`, which runs a test inside a **bubble** with a fake clock:

{{< example file="examples/m06/selects/synctest_test.go" region="synctest" run="go test -run TestTimeoutWithVirtualTime -v ./examples/m06/selects" >}}

Inside `synctest.Test`, time only advances when every goroutine in the bubble is blocked, and then jumps straight to the next timer. The 20 ms timeout fires after exactly 20 ms of virtual time, and the test finishes in microseconds of real time. `synctest.Wait()` waits until all goroutines in the bubble are blocked, which is useful to check state after background work settles. Module 9 uses it more.

## Common mistakes

### Leaking the goroutine that lost the race

A worker sending its result on an unbuffered channel blocks forever when the caller timed out and left. Give result channels a buffer of one, or make the worker also select on a done signal.

### `default` in a loop

Turns a waiting loop into a CPU-burning spin. Remove the `default` or block on something.

### Assuming `select` checks cases in order

When several cases are ready, the choice is random. If one case must take priority, check it first in a separate non-blocking `select`, or restructure the logic.

### Sleeping in tests

`time.Sleep(100 * time.Millisecond)` to "let the goroutine run" is slow and flaky. Use channels, `WaitGroup`s, or `synctest`.

## Key takeaways

- `select` blocks until one case can proceed, choosing randomly among ready cases.
- Combine an operation with `time.After` or a `context` to add a timeout, and make sure the losing goroutine can still exit.
- `default` makes a `select` non-blocking; a nil channel disables a case.
- Since Go 1.23, unreferenced timers are collected and timer channels are synchronous; still `Stop` timers and avoid `time.After` in hot loops.
- `testing/synctest` (Go 1.25+) gives tests a fake clock for deterministic timeout testing.

## Further reading

- [Select statements](https://go.dev/ref/spec#Select_statements) in the spec
- [Go 1.23 release notes: timer changes](https://go.dev/doc/go1.23#timer-changes)
- [Testing concurrent code with testing/synctest](https://go.dev/blog/synctest) on the Go blog
- [Package testing/synctest](https://pkg.go.dev/testing/synctest)
