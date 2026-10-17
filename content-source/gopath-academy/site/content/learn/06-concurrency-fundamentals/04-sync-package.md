---
title: "6.4 sync.WaitGroup, WaitGroup.Go, Mutex, RWMutex and Once"
linkTitle: "6.4 The sync package"
weight: 4
description: "Guard shared state with Mutex and RWMutex, wait for goroutines with WaitGroup and WaitGroup.Go, and initialize lazily with Once, OnceValue and OnceFunc."
goVersion: "1.27"
---

## What you will learn

- How to protect shared state with `sync.Mutex`
- When `sync.RWMutex` helps, and when it does not
- `sync.WaitGroup`, and the shorter `WaitGroup.Go` (Go 1.25+)
- Lazy, thread-safe initialization with `sync.Once`, `OnceValue` and `OnceFunc`
- When to choose a mutex over a channel

## The idea

Channels are great for passing data between goroutines. But many programs simply have **shared state** (a cache, a counter, a connection pool) that several goroutines read and modify. For that, the `sync` package's mutual exclusion locks are simpler and faster.

The rule is short: **if more than one goroutine accesses a variable and at least one of them writes, every access must be synchronized.**

## `sync.Mutex`

{{< example file="examples/m06/syncprims/syncprims_test.go" region="mutex" run="go test -race -run Example_mutex -v ./examples/m06/syncprims" >}}

- `Lock` blocks until the mutex is free; `Unlock` releases it.
- `defer mu.Unlock()` right after `Lock` guarantees the unlock even on early returns or panics.
- Put the mutex **next to the data it protects**, in the same struct, and make the fields unexported so the only way to reach them is through methods that lock.
- The zero value of a `Mutex` is an unlocked mutex. Never copy a mutex after first use; `go vet` reports copies (lesson 4.1).

Keep critical sections short: lock, touch the data, unlock. Do not hold a lock while doing I/O or calling unknown code, or every other goroutine waits behind you.

## `sync.RWMutex`

{{< example file="examples/m06/syncprims/syncprims_test.go" region="rwmutex" run="go test -race -run Example_rwMutex -v ./examples/m06/syncprims" >}}

An `RWMutex` allows **many readers at once** (`RLock`), or **one writer** (`Lock`). It helps when reads vastly outnumber writes **and** the read sections do noticeable work. For tiny critical sections it can be slower than a plain `Mutex`, because the bookkeeping costs more. Start with `Mutex`; switch to `RWMutex` when a benchmark (Module 10) shows lock contention on reads.

## `sync.WaitGroup`

You have used it since lesson 6.1. The full pattern for collecting results:

{{< example file="examples/m06/syncprims/syncprims_test.go" region="waitgroup" run="go test -race -run Example_collectWithMutex -v ./examples/m06/syncprims" >}}

`Add` must happen **before** the goroutine starts, so that `Wait` cannot return early.

### `WaitGroup.Go` (Go 1.25+)

Go 1.25 added a method that does the `Add`, `go` and `Done` for you:

{{< example file="examples/m06/syncprims/wggo_test.go" region="wggo" run="go test -race -run Example_waitGroupGo -v ./examples/m06/syncprims" >}}

It removes the two most common `WaitGroup` mistakes (forgetting `Done`, calling `Add` inside the goroutine). Prefer it in modules targeting Go 1.25 or later.

## `sync.Once`, `OnceValue`, `OnceFunc`

`sync.Once` runs a function exactly once, even if many goroutines call it at the same time, and makes everyone wait until it has finished. Go 1.21 added helpers that are more convenient:

{{< example file="examples/m06/syncprims/syncprims_test.go" region="once" run="go test -race -run Example_onceValue -v ./examples/m06/syncprims" >}}

- `sync.OnceFunc(f)` returns a function that calls `f` only the first time.
- `sync.OnceValue(f)` also caches `f`'s result.
- `sync.OnceValues(f)` caches a `(value, error)` pair.

Use them for lazy initialization of expensive, shared things: parsed templates, compiled regular expressions, a client created on first use.

## Mutex or channel?

| Situation | Prefer |
|---|---|
| Several goroutines read and update shared state (cache, counter, registry) | Mutex |
| Passing ownership of data from one goroutine to another | Channel |
| Distributing work to workers, collecting results | Channel |
| Signaling an event to many goroutines (shutdown) | Close a channel, or `context` |
| A simple counter or flag | `sync/atomic` (lesson 6.5) |

The Go wiki puts it well: use whichever is most expressive and simple for the problem. Mixing both in one design is common and fine.

## Other tools in `sync`

- `sync.Map`: a concurrent map optimized for two specific cases (keys written once and read many times, or goroutines working on disjoint keys). For everything else, a `map` plus a `Mutex` is simpler and often faster.
- `sync.Pool`: reuses temporary objects to reduce garbage collection pressure (lesson 10.6).
- `sync.Cond`: waits for a condition; rarely needed, as channels usually express the same thing more clearly.

## Common mistakes

### Forgetting to unlock on an early return

Without `defer`, any `return` or panic between `Lock` and `Unlock` leaves the mutex locked forever. Use `defer mu.Unlock()`.

### Copying a struct that contains a mutex

Passing it by value creates a second, independent lock, so the "protected" data is no longer protected. Use pointer receivers and `go vet`.

### Locking in the wrong order

Goroutine A locks `x` then `y`; goroutine B locks `y` then `x`. Each waits for the other forever. Always acquire multiple locks in the same order, or redesign to need only one.

### Returning internal maps or slices from locked methods

The caller then uses them **without** the lock. Return copies (as `Snapshot` does in the exercise).

## Exercise

{{< exercise dir="m06/parallel" >}}
Implement `Sum`, which splits a slice across worker goroutines, and `Tally`, a concurrency-safe counter whose zero value works and whose `Snapshot` returns a copy. Run the tests with `go test -race`; the race detector must stay silent.
{{< /exercise >}}

## Key takeaways

- Synchronize every access to a variable that is shared and written.
- `Mutex` with `defer Unlock`, kept next to the data it guards, is the default tool for shared state.
- `RWMutex` pays off only when reads dominate and are not trivial.
- `WaitGroup` waits for goroutines; `WaitGroup.Go` (Go 1.25+) removes common mistakes.
- `OnceValue` and friends give thread-safe lazy initialization.

## Further reading

- [Package sync](https://pkg.go.dev/sync)
- [Go wiki: Use a sync.Mutex or a channel?](https://go.dev/wiki/MutexOrChannel)
- [Go 1.25 release notes: WaitGroup.Go](https://go.dev/doc/go1.25)
