---
title: "10.4 Allocation reduction and escape analysis"
linkTitle: "10.4 Allocations"
weight: 4
description: "Understand stack versus heap, read the compiler's escape analysis with -gcflags=-m, see exactly when values and interfaces allocate, and apply the standard techniques to reduce allocations."
goVersion: "1.27"
---

## What you will learn

- Why heap allocations cost more than stack allocations
- How the compiler decides with **escape analysis**, and how to see its decisions
- Exactly when returning pointers and storing values in interfaces allocate
- Practical techniques to reduce allocations
- When to stop

## The idea

Go decides automatically whether each value lives on the **stack** (freed for free when the function returns) or the **heap** (managed by the garbage collector). Heap allocations are not slow individually, but each one costs allocator work, adds GC pressure, and hurts cache locality. In hot loops, the number of allocations often decides performance, as the benchmarks in this module have shown.

You do not choose stack or heap explicitly. The compiler's **escape analysis** proves when a value cannot outlive its function and keeps it on the stack. Your job is to write code where that proof is easy.

## Seeing escape analysis

```bash
go build -gcflags=-m ./examples/m10/escape
```

```text
examples/m10/escape/escape.go:16:45: &Point{...} escapes to heap
examples/m10/escape/escape.go:22:15: leaking param: v
examples/m10/escape/escape.go:25:12: v does not escape
```

`-gcflags=-m` prints the compiler's decisions: what escapes and why, and what was inlined. Use `-gcflags='-m -m'` for more detail on why.

## When values allocate

{{< example file="examples/m10/escape/escape.go" region="code" >}}

{{< example file="examples/m10/escape/escape_test.go" region="test" run="go test -v ./examples/m10/escape" >}}

What the test shows, on both Go 1.24 and 1.27:

- **Returning a struct by value** does not allocate. Returning a **pointer** to a local makes it escape (lesson 2.8). For small structs, values are often faster.
- **Interfaces** need their contents somewhere. Storing a non-constant `int` in an interface that **escapes** (here, saved in a global) allocates. Three cases do not:
  - small integers 0 to 255, which the runtime keeps in a static table;
  - constants, which the compiler stores in read-only data;
  - interfaces that do **not** escape, whose contents can live on the stack.
- **Fixed-size arrays** live on the stack. Recent releases (Go 1.25 and 1.26) also put more variable-sized slice backing arrays on the stack when they are small, which is one reason allocation counts can drop when you upgrade Go.

The `//go:noinline` directives make the examples show each function on its own. Inlining often removes allocations entirely: once a function is inlined into its caller, a pointer it returns may not escape the caller. Do not add `//go:noinline` to real code.

## Techniques that reduce allocations

Roughly in order of how often they help:

1. **Do not repeat expensive setup**: compile regular expressions, templates and encoders once (lesson 10.2).
2. **Preallocate** slices and maps when you know the size: `make([]T, 0, n)`, `make(map[K]V, n)`.
3. **Build strings with `strings.Builder` or `append`**, not `+=` in loops (lesson 9.7).
4. **Use `strconv.Append*` and `fmt.Appendf`** in hot formatting paths (lesson 8.6).
5. **Avoid converting between `string` and `[]byte`** repeatedly; pick one representation for the hot path.
6. **Return small structs by value**; pass pointers only when you need sharing or the struct is large.
7. **Avoid `any` in hot paths**; generic functions or concrete types avoid boxing.
8. **Reuse buffers** across iterations (`buf = buf[:0]`), or across goroutines with `sync.Pool` (lesson 10.6).
9. **Do not allocate when the common case needs nothing**: `AnalyzeFast` only calls `strings.ToLower` when a path actually contains uppercase letters.

Lock in the result with an allocation test (`testing.AllocsPerRun`, lesson 9.7), so a later change cannot quietly undo it.

## When to stop

Allocation reduction makes code harder to read. Do it in code that profiles show is hot, stop when the code meets its performance goals, and leave a comment explaining any non-obvious trick (such as the size estimate in `JoinAppend`).

## Common mistakes

### Pointers everywhere "for speed"

Passing and returning pointers to small structs often adds heap allocations. Measure before assuming.

### Trusting `-m` output from a different build

Escape decisions depend on inlining, which depends on build flags and the Go version. Check with the same toolchain and flags you ship.

### Micro-optimizing cold code

A function called once per request rarely matters. Find the loops first.

## Key takeaways

- Escape analysis keeps values on the stack when it can prove they do not outlive the function.
- `go build -gcflags=-m` shows what escapes and what was inlined.
- Returning pointers, and storing non-constant values in escaping interfaces, allocate; inlining often removes allocations.
- Preallocate, reuse buffers, avoid repeated setup and avoid unnecessary conversions in hot paths.
- Measure, lock in gains with allocation tests, and stop when it is fast enough.

## Further reading

- [Go FAQ: stack or heap?](https://go.dev/doc/faq#stack_or_heap)
- [A Guide to the Go Garbage Collector](https://go.dev/doc/gc-guide), which explains the cost model
