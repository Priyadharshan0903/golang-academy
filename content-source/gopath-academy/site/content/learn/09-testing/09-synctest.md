---
title: "9.9 Testing concurrent code with testing/synctest"
linkTitle: "9.9 testing/synctest"
weight: 9
description: "Test code with goroutines, timers and timeouts deterministically: synctest bubbles, virtual time, synctest.Wait and synctest.Sleep, and what can and cannot run inside a bubble."
goVersion: "1.27"
---

## What you will learn

- Why concurrent, time-based code is hard to test
- How `synctest.Test` runs a test in a **bubble** with a virtual clock
- What `synctest.Wait` guarantees, and `synctest.Sleep` (Go 1.27)
- The rules for what can run inside a bubble

{{< callout type="info" >}}
`testing/synctest` is available from **Go 1.25**. The examples here are gated with `//go:build go1.25`.
{{< /callout >}}

## The problem

Here is a cache whose entries expire, with a background goroutine that sweeps expired entries:

{{< example file="examples/m09/synctesting/cache.go" region="code" >}}

Testing it the obvious way means real sleeps: wait 61 seconds for an entry to expire, then a bit more for the sweeper to run. Shrink the durations to milliseconds and the test becomes **flaky**: on a busy CI machine, the sweeper might not have run yet when you check.

## The solution: a bubble with fake time

{{< example file="examples/m09/synctesting/cache_test.go" region="test" run="go test -v ./examples/m09/synctesting" >}}

Inside `synctest.Test(t, func(t *testing.T) { ... })`:

- **Time is virtual.** It starts at a fixed instant and only advances when **every goroutine in the bubble is blocked**. Then it jumps straight to the next timer. So `time.Sleep(59 * time.Second)` returns immediately in real time, but `time.Now()` afterwards is exactly 59 seconds later.
- **Goroutines started inside belong to the bubble**, including the cache's sweeper and its ticker.
- **`synctest.Wait()`** blocks until every other goroutine in the bubble is durably blocked. After `Wait`, you know background work triggered by the last time step has finished. No polling, no sleeps.
- Go 1.27 adds **`synctest.Sleep(d)`**, which combines `time.Sleep(d)` and `synctest.Wait()`, the most common pairing.
- After the function returns, `Test` waits for every goroutine in the bubble to exit. If they are all blocked with no way to proceed, the test fails with a deadlock report, which catches leaked goroutines.

The whole test runs in microseconds and is fully deterministic: same result every time, on any machine.

## What "blocked" means

Time advances only when goroutines are **durably blocked**, meaning blocked on something only another goroutine in the bubble can unblock:

- channel operations on channels created in the bubble;
- `sync.WaitGroup.Wait`, `sync.Cond.Wait`, `select` with no ready case;
- `time.Sleep` and timers.

Blocking on **I/O** (a real network socket, a file) or on a `sync.Mutex` held by another goroutine does **not** count, because something outside the bubble could unblock it. That is why Go 1.27's `httptest.NewTestServer` uses an in-memory network: HTTP clients and servers then block on in-bubble operations, and timeouts become testable with virtual time.

## Rules and limits

- Channels, timers and tickers created inside a bubble must not be used from outside it, and vice versa.
- Do not call `t.Run`, `t.Parallel` or `t.Deadline` on the `*testing.T` passed into the bubble.
- Code that waits on real I/O will not advance virtual time; use fakes or the in-memory network.

## Common mistakes

### Mixing real and virtual time

Passing a timer or channel created outside the bubble into it panics or deadlocks. Create everything the test needs inside the function.

### Forgetting `Wait` before asserting

After `time.Sleep`, the goroutines woken by that time step may not have run yet. Call `synctest.Wait()` (or `synctest.Sleep`) before checking their effects.

### Leaving goroutines running

`Test` waits for every goroutine in the bubble to exit. Stop background workers (`defer c.Close()`) before the function returns, or the test reports a deadlock.

## Key takeaways

- `synctest.Test` runs a test in a bubble with a virtual clock that advances only when all bubble goroutines are blocked.
- Sleeps and timeouts become instant and deterministic.
- `synctest.Wait` waits for background goroutines to settle; Go 1.27's `synctest.Sleep` sleeps and waits.
- I/O does not count as blocked; use fakes or `httptest.NewTestServer` (Go 1.27) for network code.

## Further reading

- [Testing concurrent code with testing/synctest](https://go.dev/blog/synctest) on the Go blog
- [Package testing/synctest](https://pkg.go.dev/testing/synctest)

## Module 9 recap

You can now write table-driven tests with good failure messages, test your tests with planted bugs, use helpers and golden files, document APIs with examples, isolate dependencies with fakes, test HTTP code with `httptest`, fuzz parsers, benchmark with `b.Loop`, measure coverage honestly, and test concurrent code deterministically with `synctest`.

Next is **[Module 10: Profiling and performance](/learn/10-performance/)**.
