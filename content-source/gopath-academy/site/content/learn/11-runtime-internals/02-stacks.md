---
title: "11.2 Stacks and goroutine growth"
linkTitle: "11.2 Goroutine stacks"
weight: 2
description: "Why goroutines are cheap: small initial stacks that grow by copying, what that means for pointers and unsafe code, how deep recursion behaves, and what a stack overflow looks like."
goVersion: "1.27"
---

## What you will learn

- How big a goroutine's stack starts, and how it grows and shrinks
- Why stacks **move**, and what that means for addresses
- How deep recursion behaves in Go, and what the limit is
- How much memory a blocked goroutine really costs

## The idea

An operating system thread typically gets a fixed stack of 1 to 8 MiB, reserved up front. That is why programs cannot have a million threads. A goroutine starts with a tiny stack, **2 KiB** at minimum, and the runtime grows it only when needed. That is the main reason goroutines are cheap enough to create one per request, per connection or per task.

## Measuring the cost

This function starts `n` goroutines that block on a channel, and measures how much stack memory they added:

{{< example file="examples/m11/stacks/stacks.go" region="cost" run="go test -v -run TestStackCost ./examples/m11/stacks" >}}

On both Go 1.24 and Go 1.27 it reports `about 2048 bytes of stack per blocked goroutine`. Ten thousand idle goroutines cost about 20 MiB of stack, plus a few hundred bytes each for the runtime's goroutine record. A thread per connection would need gigabytes of reserved address space for the same.

Since Go 1.19, the runtime adapts the starting size: if goroutines in your program typically need more than 2 KiB, new ones start with the average observed size, which avoids repeated growth. Goroutines like the ones above, which use very little, still start at the minimum.

## How stacks grow

Every Go function starts with a short **prologue** that compares the stack pointer with the end of the current stack. If the function's frame does not fit, it calls into the runtime, which:

1. allocates a new stack **twice as large**,
2. **copies** the old stack into it,
3. **adjusts every pointer** that points into the old stack, using the compiler's precise information about where pointers live in each frame,
4. frees the old stack and resumes the function.

Go used "segmented" stacks, linked chunks, before Go 1.3, but a function call in a loop right at a segment boundary kept allocating and freeing segments (the "hot split" problem). Contiguous, copied stacks avoid it: growth is amortized, like `append`.

Stacks also **shrink**: during garbage collection, a goroutine using less than a quarter of its stack gets it copied into one half the size.

## Stacks move

Because growth copies the stack, the **address** of a variable on the stack can change while the program runs:

{{< example file="examples/m11/stacks/stacks.go" region="moves" run="go test -v -run TestStackMoves ./examples/m11/stacks" >}}

On the test VM with Go 1.27:

```text
stacks_test.go:20: before 0x369b47048f88, after 0x369b474bff88
```

Same variable `x`, different address after the recursion grew the stack. Ordinary Go code never notices, because the runtime updates every real pointer. But a `uintptr` is just a number, and the runtime does not update numbers. That is why the `unsafe` rules (lesson 11.4) forbid keeping an address in a `uintptr` and converting it back to a pointer later: by then it may point to freed memory.

Heap objects, by contrast, never move in Go's current garbage collector.

## Deep recursion

Go has no tail-call optimization, so every recursive call uses a new frame. With growable stacks that is usually fine:

{{< example file="examples/m11/stacks/stacks.go" region="sum" >}}

`Sum(1_000_000)` recurses a million levels and needs tens of MiB of stack; the test runs it in a fresh goroutine and it completes in about 40 ms. Deep recursion that would crash a C program with a fixed 8 MiB stack works in Go.

There is a limit: by default **1 GB** on 64-bit systems (250 MB on 32-bit). Infinite recursion hits it and the program dies:

```text
runtime: goroutine stack exceeds 1000000000-byte limit
fatal error: stack overflow
```

This is a **fatal error**, not a panic: `recover` cannot catch it and the whole process exits, printing the repeating stack trace that shows where the recursion is. `debug.SetMaxStack` changes the limit, which is occasionally useful to make runaway recursion fail sooner in tests.

## Common mistakes

### Unbounded recursion on user input

Parsing deeply nested JSON or expressions recursively can be driven to the 1 GB limit by a malicious input, crashing the process. Limit the depth (the standard library's decoders do), or use an explicit stack.

### Keeping addresses in uintptr

The address may be stale after the next stack growth. Keep real pointers.

### Fearing goroutines for their memory

An idle goroutine costs about 2 KiB of stack plus a small record. The real costs of goroutines are what they hold on to and the work they do, and leaking them (lesson 7.6).

## Key takeaways

- Goroutine stacks start at 2 KiB (or the program's observed average) and grow by doubling and copying.
- Stack growth moves variables; the runtime fixes real pointers, never `uintptr` values.
- Stacks shrink during GC when mostly unused.
- Deep recursion works, up to a 1 GB default limit; exceeding it is a fatal, unrecoverable error.

## Further reading

- [Contiguous stacks design document](https://golang.org/s/contigstacks)
- [Package runtime/debug: SetMaxStack](https://pkg.go.dev/runtime/debug#SetMaxStack)
