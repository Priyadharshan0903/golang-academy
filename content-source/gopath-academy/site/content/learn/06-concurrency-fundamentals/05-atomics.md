---
title: "6.5 sync/atomic and atomic types"
linkTitle: "6.5 sync/atomic"
weight: 5
description: "Lock-free counters, flags, compare-and-swap and atomically swapped snapshots with the typed atomics in sync/atomic."
goVersion: "1.27"
---

## What you will learn

- What atomic operations are, and when they beat a mutex
- The typed atomics: `atomic.Int64`, `atomic.Bool`, `atomic.Pointer[T]` and friends
- Compare-and-swap for "exactly one winner" logic
- Publishing immutable snapshots with `atomic.Pointer`
- Why atomics are easy to misuse for anything bigger than one value

## The idea

An **atomic operation** reads or modifies a single value in one indivisible step. No other goroutine can see it half done, and no lock is needed. The CPU provides these operations directly, so they are very fast.

Atomics are perfect for **one independent value**: a request counter, a "shutting down" flag, a pointer to the current configuration. As soon as two values must change together consistently, use a mutex.

## Typed atomics

Since Go 1.19, `sync/atomic` has types whose methods are always atomic and whose zero values are ready to use:

{{< example file="examples/m06/atomics/atomics_test.go" region="counter" run="go test -race -run Example_atomicTypes -v ./examples/m06/atomics" >}}

| Type | Methods |
|---|---|
| `atomic.Int32`, `Int64`, `Uint32`, `Uint64`, `Uintptr` | `Load`, `Store`, `Add`, `Swap`, `CompareAndSwap` (plus `And`, `Or` since Go 1.23) |
| `atomic.Bool` | `Load`, `Store`, `Swap`, `CompareAndSwap` |
| `atomic.Pointer[T]` | `Load`, `Store`, `Swap`, `CompareAndSwap` |
| `atomic.Value` | `Load`, `Store` for any type (older; prefer `Pointer[T]`) |

Prefer these types over the older functions like `atomic.AddInt64(&n, 1)`. With the functions, it is easy to forget to use them on **every** access; with the types, there is no other way to touch the value. `int64` atomics also need 64-bit alignment on 32-bit platforms, which the types guarantee for you.

## Compare-and-swap

`CompareAndSwap(old, new)` sets the value to `new` **only if** it currently equals `old`, and reports whether it did. It is the building block of lock-free algorithms:

{{< example file="examples/m06/atomics/atomics_test.go" region="cas" run="go test -race -run Example_compareAndSwap -v ./examples/m06/atomics" >}}

Exactly one goroutine wins the race to claim ownership, with no lock.

## Atomic snapshots with `atomic.Pointer`

A powerful pattern for read-heavy shared data: keep it **immutable**, and atomically swap a pointer to a new version when it changes.

{{< example file="examples/m06/atomics/atomics_test.go" region="pointer" run="go test -race -run Example_atomicPointer -v ./examples/m06/atomics" >}}

Readers call `Load` and get a consistent snapshot with no locking at all. A writer builds a **new** `Settings` and stores the pointer. Old readers keep using the old snapshot until they finish. This is how many services implement hot-reloadable configuration and routing tables.

The rule that makes it work: **never modify a value after publishing it.** Build a new one.

## When not to use atomics

- **Several related values.** Updating `count` and `total` with two atomic operations lets a reader see the new `count` with the old `total`. Use a mutex.
- **Check then act.** `if n.Load() < limit { n.Add(1) }` is not atomic as a whole; two goroutines can both pass the check. Use `CompareAndSwap` in a loop, or a mutex.
- **Clarity.** A mutex is easier to reason about. Reach for atomics when profiling shows lock contention, or for the clear single-value cases above.

## Common mistakes

### Mixing atomic and plain access

Using `atomic.AddInt64(&n, 1)` in one place and `n++` in another is a data race. The typed atomics prevent this.

### Assuming atomics make compound operations safe

Two atomic operations in a row are not one atomic operation.

### Mutating a published snapshot

Changing a field on a struct reachable through `atomic.Pointer` is a data race with every reader. Copy, modify the copy, store the new pointer.

## Key takeaways

- Atomics update a single value indivisibly and without locks.
- Use the typed atomics (`atomic.Int64`, `atomic.Bool`, `atomic.Pointer[T]`); their zero values are ready.
- `CompareAndSwap` enables "exactly one winner" logic.
- Publish immutable snapshots with `atomic.Pointer` for lock-free reads.
- For anything involving more than one value, use a mutex.

## Further reading

- [Package sync/atomic](https://pkg.go.dev/sync/atomic)
- [The Go Memory Model: Atomic values](https://go.dev/ref/mem#atomic)
