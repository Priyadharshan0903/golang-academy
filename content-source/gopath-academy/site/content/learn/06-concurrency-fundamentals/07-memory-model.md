---
title: "6.7 The Go memory model, in plain words"
linkTitle: "6.7 Memory model"
weight: 7
description: "What the Go memory model guarantees, what happens-before means, which operations create it, and why unsynchronized flags and double-checked locking fail."
goVersion: "1.27"
---

## What you will learn

- Why a write in one goroutine is not automatically visible in another
- What **happens before** means
- Which operations create happens-before relationships
- Why "spin until a flag is set" is broken
- The one-sentence rule that keeps you safe

## The idea

Modern CPUs and compilers reorder and cache memory operations to run faster. Within a single goroutine, this is invisible: your code behaves as if it ran in order. **Across goroutines, it is not.** Without synchronization, one goroutine may see another's writes late, out of order, or never.

The Go memory model is the contract that tells you when a write is **guaranteed** to be visible to a read in another goroutine. The guarantee is called **happens before**.

## Happens before

If event A **happens before** event B, then B sees the effects of A (and of everything that happened before A).

Within one goroutine, statements happen before each other in program order. **Between goroutines**, happens-before only exists when it is created by a synchronizing operation. The ones you will use:

| Operation | Guarantee |
|---|---|
| `go f()` | The `go` statement happens before `f` starts running |
| Channel send | A send happens before the corresponding receive completes |
| Channel close | A close happens before a receive that returns because the channel is closed |
| Unbuffered receive | The receive happens before the corresponding send completes |
| `Mutex.Unlock` | Unlocking happens before any later `Lock` returns |
| `WaitGroup.Done` | Each `Done` happens before the `Wait` it unblocks returns |
| `Once.Do(f)` | `f` returning happens before any `Do` call returns |
| Atomic operations | An atomic write is visible to atomic reads that observe it, like a sequentially consistent memory |

## A correct hand-off

{{< example file="examples/m06/memorymodel/memorymodel_test.go" region="handoff" run="go test -race -run Example_channelHandoff -v ./examples/m06/memorymodel" >}}

Everything the goroutine wrote to `r` before sending is visible to the receiver, because the send happens before the receive completes. After the send, the goroutine must not touch `r` again: ownership has moved to the receiver.

## A broken flag, and the fix

{{< example file="examples/m06/memorymodel/memorymodel_test.go" region="broken" run="go test -race -run Example_fixedWithWaitGroup -v ./examples/m06/memorymodel" >}}

In the broken version (shown in the comment), nothing synchronizes `ready`. The compiler is allowed to hoist the read of `ready` out of the loop, so it may spin forever; and even if it sees `ready == true`, it is not guaranteed to see `msg`. Real programs have hung exactly like this. The fix is any synchronizing operation from the table; a `WaitGroup` is the simplest here.

The same reasoning breaks "double-checked locking" (checking a plain variable before taking a lock to skip the lock). In Go, use `sync.Once` or `sync.OnceValue` instead.

## The rule

The memory model document itself says that if you need its details to understand your program, you are being too clever, and ends that advice with: "Don't be clever."

In practice: **every variable shared between goroutines and written by any of them must be accessed only through a channel, a mutex, an atomic, or another `sync` primitive.** Run `-race` to check yourself.

## What Go guarantees even with races

Go is more forgiving than C or C++ in one respect: a racy program cannot corrupt memory in arbitrary ways for **single-word** values; a read sees some value that was actually written. But multiword values like strings, slices and interfaces can be **torn** (you see the pointer from one write and the length from another), which can crash the program or read out of bounds. Do not rely on any of this. Synchronize.

## Common mistakes

### Unsynchronized "done" flags

`for !done {}` or `if !initialized { ... }` with plain booleans. Use a channel, `atomic.Bool`, or `sync.Once`.

### Publishing a pointer before the data is complete

Storing a pointer in a shared variable and then filling in its fields. Readers may see the pointer before the fields. Build completely, then publish through a synchronizing operation.

### Assuming `time.Sleep` synchronizes

Sleeping makes a race less likely to show up; it creates no happens-before relationship.

## Key takeaways

- Without synchronization, writes in one goroutine may never be seen by another.
- Happens-before is created by `go`, channel operations, mutexes, `WaitGroup`, `Once` and atomics.
- Hand off data through channels and stop touching it after sending.
- Never spin on plain flags; never use double-checked locking; use `sync.Once`.
- Do not be clever: synchronize every shared, written variable, and run `-race`.

## Further reading

- [The Go Memory Model](https://go.dev/ref/mem)
- [Updating the Go Memory Model](https://research.swtch.com/gomm) by Russ Cox
