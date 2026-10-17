---
title: "7.6 Goroutine leaks: finding and preventing them"
linkTitle: "7.6 Goroutine leaks"
weight: 6
description: "Why goroutines leak, the patterns that cause most leaks, how to detect them in tests and production, and the goroutineleak profile in Go 1.27."
goVersion: "1.27"
---

## What you will learn

- What a goroutine leak is and why it matters
- The handful of patterns that cause almost every leak
- How to test for leaks
- How to find leaks in a running service with `pprof`
- The `goroutineleak` profile, new in Go 1.27

## The idea

A goroutine **leaks** when it can never finish: it is blocked forever on a channel, a lock or I/O that will never complete. The garbage collector does not free goroutines that are blocked, nor anything they reference. Each leak is small, but leaks usually happen **per request**, so memory and goroutine counts grow steadily until the service is restarted or killed.

## The usual causes

Almost every leak is one of these:

1. **A send nobody receives.** The receiver returned early (timeout, error, first result won).
2. **A receive nobody sends to, or a `range` over a channel nobody closes.**
3. **A goroutine waiting on a context nobody cancels**, or not watching any context at all.
4. **A lock that is never unlocked.**
5. **Blocking I/O without a timeout**, such as an HTTP request with no deadline.

Here is cause 1 in its most common form, "ask several, take the fastest":

{{< example file="examples/m07/leaks/leaks_test.go" region="leaky" >}}

After the first value is received, every other replica blocks on its send forever. The fix gives the channel enough buffer for every send, so all goroutines can finish. The other standard fix is to make the sender `select` on `ctx.Done()`.

## Prevent leaks by design

{{< example file="examples/m07/leaks/leaks_test.go" region="ctxfix" run="go test -run Example_contextPreventsLeak -v ./examples/m07/leaks" >}}

Rules that make leaks rare:

- **Every goroutine has an owner** who knows how it stops.
- **Every blocking operation has a way out**: a `ctx.Done()` case, a closed channel, a timeout.
- **Result channels that may be abandoned have a buffer** big enough for all senders.
- **`defer cancel()`** right after creating a context.

## Detecting leaks in tests

The simplest check counts goroutines before and after:

{{< example file="examples/m07/leaks/leaks_test.go" region="detect" run="go test -v -run Version ./examples/m07/leaks" >}}

Because goroutines take a moment to exit, poll for a short time rather than checking once. The exercises in Modules 6 and 7 use exactly this technique. For larger projects, `go.uber.org/goleak` does it more thoroughly: it lists the leaked goroutines' stacks and ignores known background goroutines.

## Finding leaks in production

Import `net/http/pprof` in your service (on an internal port only), and you get:

```bash
# How many goroutines, grouped by stack:
curl -s http://localhost:6060/debug/pprof/goroutine?debug=1 | head -40

# Or explore interactively:
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

A leak shows up as a stack with an ever-growing count, all blocked at the same line (`chan send`, `chan receive`, `select`). The `runtime/metrics` value `/sched/goroutines` (Go 1.26+) is a good thing to graph and alert on.

## The `goroutineleak` profile (Go 1.27)

The goroutine profile shows **every** blocked goroutine, and many of them are blocked legitimately (waiting for the next request, say). Go 1.26 introduced, as an experiment, and Go 1.27 made generally available, a profile that reports only goroutines the runtime can **prove** are leaked: goroutines blocked on a channel, mutex or condition that no running goroutine can reach anymore. The garbage collector finds them during its reachability analysis.

{{< example file="examples/m07/leaks/leakprofile_test.go" region="profile" run="go test -v -run Profile ./examples/m07/leaks" >}}

In a service it is served at `/debug/pprof/goroutineleak` by `net/http/pprof`. It cannot find every leak (a goroutine blocked on a channel that is still referenced somewhere is not provably leaked), but what it reports is almost certainly a bug.

## Common mistakes

### Assuming the garbage collector will clean up

It does not free blocked goroutines. A leaked goroutine lives until the process exits.

### Timeouts on the caller side only

`select` with `time.After` protects the caller, but if the worker's send is unbuffered, the worker leaks. Fix both sides.

### HTTP clients without timeouts

`http.Get` with the default client has **no** timeout. A hung server leaks the goroutine (and the connection) forever. Use a client with `Timeout`, or a request context with a deadline (lesson 8.7).

## Key takeaways

- A leaked goroutine is blocked forever and is never freed, along with everything it references.
- Most leaks are unreceived sends, unclosed channels, uncancelled contexts or I/O without timeouts.
- Give every goroutine an owner and every blocking operation an exit.
- Test for leaks by comparing goroutine counts, or with `goleak`.
- In production, watch the goroutine profile and metrics; on Go 1.27, check `/debug/pprof/goroutineleak`.

## Further reading

- [Go 1.27 release notes: goroutineleak profile](https://go.dev/doc/go1.27)
- [Package runtime/pprof](https://pkg.go.dev/runtime/pprof) and [net/http/pprof](https://pkg.go.dev/net/http/pprof)
- [go.uber.org/goleak](https://pkg.go.dev/go.uber.org/goleak)
