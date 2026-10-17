---
title: "11.4 Memory layout, alignment and unsafe"
linkTitle: "11.4 Memory layout"
weight: 4
description: "How Go lays out values in memory: sizes, alignment and padding, the headers behind strings, slices and interfaces, false sharing between CPU cores, and how to use the unsafe package without breaking your program."
goVersion: "1.27"
---

## What you will learn

- The size and alignment of Go's types, and where struct padding comes from
- How to shrink a struct by reordering its fields
- What strings, slices, interfaces and maps look like in memory
- **False sharing**, and a measured 5x slowdown it causes
- What the `unsafe` package can do, and the rules for using it safely

## Sizes and alignment

Every type has a **size** and an **alignment**. A value's address must be a multiple of its alignment, because CPUs read aligned data most efficiently (and some cannot read misaligned data at all). On 64-bit platforms:

| Type | Size | Alignment |
|---|---|---|
| `bool`, `int8`, `uint8` | 1 | 1 |
| `int16` | 2 | 2 |
| `int32`, `float32`, `rune` | 4 | 4 |
| `int64`, `float64`, `int`, `uintptr`, pointers | 8 | 8 |
| `string` | 16 | 8 |
| slice | 24 | 8 |
| interface | 16 | 8 |
| `struct{}` | 0 | 1 |

A struct's fields are laid out **in the order you declare them**. The compiler inserts **padding** before a field to align it, and at the end so the struct's size is a multiple of its alignment (so arrays of it stay aligned). The `unsafe` package shows the result:

{{< example file="examples/m11/layout/layout.go" region="structs" >}}

{{< example file="examples/m11/layout/layout_test.go" region="example" run="go test -run Example -v ./examples/m11/layout" >}}

`Wasteful` and `Compact` hold the same five fields, but `Wasteful` is **twice as big**: 32 bytes, of which 17 are padding. Ordering fields from largest alignment to smallest removes most padding. For one value, 16 bytes do not matter. For ten million values in a slice or cache, that is 160 MB.

The `fieldalignment` analyzer finds such structs for you:

```bash
go run golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest ./...
```

Do not reorder blindly, though. Field order is also documentation (related fields together), and a struct used once per request is not worth making less readable. Reorder the types you hold many of.

## What the headers tell you

The second example shows the sizes of Go's built-in "header" types:

- A **string** is a pointer to bytes plus a length (16 bytes). Assigning or passing a string copies only those 16 bytes, never the text.
- A **slice** is a pointer, a length and a capacity (24 bytes, lesson 3.2).
- An **interface** is two words: type information and a pointer to the value (16 bytes, lesson 11.5).
- A **map**, **channel** or **function** value is a single pointer (8 bytes) to a runtime structure. This is why passing a map to a function lets the function modify the caller's map.
- `struct{}` takes **zero** bytes, which makes `map[string]struct{}` the idiomatic set and `chan struct{}` the idiomatic signal.

## False sharing

CPUs move memory between main memory and their caches in **cache lines**, usually 64 bytes. When two cores write to different variables that happen to share a cache line, each write invalidates the line in the other core's cache, and the line bounces between cores. This is **false sharing**: no data is actually shared, but the hardware behaves as if it were.

{{< example file="examples/m11/layout/layout.go" region="sharing" >}}

Two goroutines each increment their **own** counter a million times:

{{< example file="examples/m11/layout/layout_test.go" region="bench" >}}

On the 2-core test VM with Go 1.27:

```text
           │   adjacent   │               padded                │
           │    sec/op    │    sec/op     vs base               │
Counters-2   49.021m ± 8%   9.110m ± 11%  -81.42% (p=0.002 n=6)
```

The only difference is 120 bytes of padding: the padded version is about **5 times faster**. False sharing matters for per-core or per-worker counters, statistics and queue indices updated in hot loops. The standard library and runtime pad such structures; the `golang.org/x/sys/cpu` package provides a `CacheLinePad` type sized for the platform.

## The unsafe package

`unsafe` lets you step outside Go's type system. It is used inside the standard library and the runtime, and in a few performance-critical libraries. Ordinary application code should almost never import it. Its parts:

- `unsafe.Sizeof`, `Alignof` and `Offsetof`: compile-time constants about layout. These are safe.
- `unsafe.Pointer`: a pointer that can be converted to and from any pointer type, and to `uintptr`.
- `unsafe.String`, `unsafe.StringData`, `unsafe.Slice`, `unsafe.SliceData` (Go 1.17 and 1.20): build strings and slices from raw pointers without copying.

The danger is real. This converts a byte slice to a string without copying:

{{< example file="examples/m11/layout/layout_test.go" region="unsafe" >}}

The output is `jello`: a Go string **changed**, which the language promises can never happen. Code elsewhere that used `s` as a map key, or assumed it was immutable, is now silently broken. Zero-copy conversion is only correct if the bytes are **never modified again**, and nothing in the language checks that for you.

The rules for `unsafe.Pointer`, from the package documentation, allow a small set of patterns, for example converting `*T1` to `*T2` when the memory layouts are compatible, or doing pointer arithmetic with `uintptr` **within a single expression**. Anything else is invalid even if it seems to work. Two consequences you saw in this module:

- Never store a `uintptr` and convert it back later: the object may have moved (stacks, lesson 11.2) or been freed.
- Invalid `unsafe` code can break when the compiler improves. Go 1.25's release notes warn that allocating more slices on the stack can expose incorrect `unsafe.Pointer` use that previously appeared to work.

`go vet` checks some misuses (its `unsafeptr` check), and `-gcflags=all=-d=checkptr` (enabled automatically by `-race`) checks others at run time.

## Common mistakes

### Padding-heavy structs held by the million

Order large-alignment fields first in types you store in bulk.

### Per-goroutine counters packed into one array

`counts [8]int64` written by 8 goroutines puts them all in one or two cache lines. Pad each element, or let each goroutine accumulate locally and combine at the end (usually best).

### unsafe for convenience

If you reach for `unsafe` to avoid a copy, first measure that the copy matters. It usually does not.

## Key takeaways

- Fields are laid out in declaration order with padding for alignment; largest-alignment-first ordering minimizes size.
- Strings are 16 bytes, slices 24, interfaces 16; maps, channels and functions are pointers; `struct{}` is free.
- False sharing makes independent writes from different cores slow; padding to separate cache lines fixed a 5x slowdown here.
- `unsafe` breaks Go's guarantees; follow the documented patterns exactly, and prefer not to use it.

## Further reading

- [Package unsafe](https://pkg.go.dev/unsafe), including the rules for valid `unsafe.Pointer` use
- [The Go spec: size and alignment guarantees](https://go.dev/ref/spec#Size_and_alignment_guarantees)
- [fieldalignment analyzer](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/fieldalignment)
