---
title: "10.6 sync.Pool and when it helps"
linkTitle: "10.6 sync.Pool"
weight: 6
description: "Reuse temporary objects across goroutines with sync.Pool, avoid its pitfalls (stale data, retained huge buffers, values instead of pointers), and measure whether it helps."
goVersion: "1.27"
---

## What you will learn

- What `sync.Pool` is for, and how it interacts with the garbage collector
- The correct get, reset, use, copy out, put pattern
- The pitfalls: stale data, huge buffers, storing values instead of pointers
- A measured comparison, and when a pool is not worth it

## The idea

Some code allocates the same kind of **temporary** object over and over: a buffer to format a log line, a scratch slice to parse a request, an encoder state. `sync.Pool` keeps such objects around so they can be **reused** instead of allocated again.

Important properties:

- A pool is safe for concurrent use and scales well: each processor has its own local cache.
- The garbage collector may **empty pools at any time** (over two GC cycles). A pool is a cache of temporary objects, never storage for anything you need to keep.
- `Get` returns an arbitrary pooled object, or calls `New` if the pool is empty.

## The pattern

Without a pool:

{{< example file="examples/m10/pool/pool.go" region="nopool" >}}

With a pool:

{{< example file="examples/m10/pool/pool.go" region="pool" >}}

The steps, each of which matters:

1. **Get and type-assert.** Store **pointers** in the pool (`*bytes.Buffer`). Putting a non-pointer value into a `sync.Pool` allocates when it is converted to `any`, which defeats the purpose (`go vet` and staticcheck can flag this).
2. **Reset.** A pooled object still contains whatever the previous user left in it.
3. **Use it.**
4. **Copy out results** before putting the object back. `buf.String()` copies; returning `buf.Bytes()` would hand the caller memory that another goroutine is about to overwrite.
5. **Put it back, unless it grew huge.** One request that needed a 10 MB buffer should not leave 10 MB buffers in the pool forever.

## Measured

`BenchmarkRender` runs both versions in parallel on all cores. Results from the 2-core test VM, Go 1.24, 6 runs:

```text
         │    nopool    │                pool                 │
         │    sec/op    │    sec/op     vs base               │
Render-2   79.83n ± 16%   56.85n ± 28%  -28.79% (p=0.004 n=6)

         │   nopool    │               pool                │
         │    B/op     │    B/op     vs base               │
Render-2   112.00 ± 0%   48.00 ± 0%  -57.14% (p=0.002 n=6)

         │   nopool   │               pool                │
         │ allocs/op  │ allocs/op   vs base               │
Render-2   2.000 ± 0%   1.000 ± 0%  -50.00% (p=0.002 n=6)
```

The pool removed the buffer allocation; the one remaining allocation is the result string, which must be a new value. The time saving is real but modest, and the `±` values show how noisy such small measurements are.

## When a pool is worth it

Use `sync.Pool` when **all** of these hold:

- the objects are temporary and expensive enough to matter (buffers, encoders, large scratch structures);
- a profile shows their allocation is significant;
- they are allocated very frequently, from many goroutines.

The standard library uses pools exactly this way, for example in `fmt` for its print state and in `encoding/json` for encode buffers.

Do **not** use a pool for small objects (a few bytes: the bookkeeping costs more than the allocation), for objects with expensive or error-prone reset logic, or as a connection pool (connections need lifecycle management; use `database/sql` or the HTTP transport's own pooling).

## Common mistakes

### Forgetting `Reset`

Output from one request leaks into the next, which is both a bug and potentially a security issue.

### Using an object after `Put`

Another goroutine may already be writing into it. Copy out first; never keep references.

### Pooling values instead of pointers

Every `Put` allocates. Pool `*T`.

### Retaining huge objects

Cap the size of objects you put back.

## Key takeaways

- `sync.Pool` reuses temporary objects across goroutines; the GC may drop them at any time.
- Pattern: Get, Reset, use, copy results out, Put (unless oversized). Pool pointers.
- Use it for frequent, expensive temporaries that a profile shows matter; measure the result.

## Further reading

- [Package sync: Pool](https://pkg.go.dev/sync#Pool)
