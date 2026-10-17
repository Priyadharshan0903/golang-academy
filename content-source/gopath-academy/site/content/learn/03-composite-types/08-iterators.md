---
title: "3.8 Iterators and range-over-func"
linkTitle: "3.8 Iterators"
weight: 8
description: "Write your own iterators with iter.Seq and iter.Seq2, range over them like slices, compose lazy pipelines, and use the iterator functions in the standard library."
goVersion: "1.27"
---

## What you will learn

- What a Go iterator is: a function that pushes values to a `yield` callback
- How `for range` works over functions (Go 1.23+)
- How early `break` is communicated back to the iterator
- How to compose lazy pipelines like filter and map
- The iterators in `slices`, `maps` and `strings`, and pull iterators with `iter.Pull`

## The idea

Since Go 1.23, `for range` works not only over slices, maps, strings, channels and integers, but also over **functions** of a specific shape. A function that takes a `yield` callback and calls it once per value is an **iterator**:

```go
type Seq[V any]     func(yield func(V) bool)
type Seq2[K, V any] func(yield func(K, V) bool)
```

Those two types live in the `iter` package. (`[V any]` means they are generic, which Module 4 explains; for now read `iter.Seq[int]` as "an iterator of ints".)

Iterators let you expose sequences of values **without building a slice first**: walking a tree, reading database rows, generating numbers, paging through an API. Values are produced on demand.

## Writing and using an iterator

{{< example file="examples/m03/iterators/iterators_test.go" region="seq" run="go test -run Example_rangeOverFunc -v ./examples/m03/iterators" >}}

How it works:

1. `for n := range Countdown(3)` calls the iterator function, passing a `yield` function that the compiler builds from the loop body.
2. Each `yield(i)` runs the loop body once with `n = i`.
3. `yield` returns `true` to ask for more, or `false` when the loop body executed `break` (or `return`).
4. The iterator **must stop** when `yield` returns `false`. Calling `yield` again after that panics.

The second loop shows the early stop: `Countdown(100)` only produces 100, 99, 98 and then returns, because the loop broke. It never counts all the way down.

## Pairs: `iter.Seq2`

`Seq2` yields two values per step, like ranging over a slice (index, value) or a map (key, value):

{{< example file="examples/m03/iterators/iterators_test.go" region="seq2" run="go test -run Example_seq2 -v ./examples/m03/iterators" >}}

## Composing lazy pipelines

Because iterators are just functions, you can wrap one in another. Nothing runs until someone ranges over the outermost iterator:

{{< example file="examples/m03/iterators/iterators_test.go" region="compose" run="go test -run Example_composition -v ./examples/m03/iterators" >}}

Note the line `if keep(v) && !yield(v) { return }`. Forwarding the stop signal from the inner `yield` is what makes the whole pipeline stop early when the consumer breaks.

`slices.Collect` drains an iterator into a slice. It is the bridge back to ordinary slices.

## Iterators in the standard library

{{< example file="examples/m03/iterators/iterators_test.go" region="stdlib" run="go test -run Example_standardIterators -v ./examples/m03/iterators" >}}

| Function | Yields |
|---|---|
| `slices.All(s)` | index, value |
| `slices.Values(s)` | values |
| `slices.Backward(s)` | index, value from the end |
| `maps.Keys(m)`, `maps.Values(m)`, `maps.All(m)` | keys, values, pairs |
| `slices.Collect(seq)`, `slices.Sorted(seq)` | (consume an iterator into a slice) |
| `maps.Collect(seq2)` | (consume pairs into a map) |
| `strings.Lines(s)` | each line, including its newline (Go 1.24+) |
| `strings.SplitSeq(s, sep)`, `strings.FieldsSeq(s)` | split pieces without allocating a slice (Go 1.24+) |

## Pull iterators

The iterators so far are **push** iterators: the iterator drives the loop. Sometimes you need to **pull** values one at a time, for example to merge two sorted sequences. `iter.Pull` converts a push iterator into a `next` function and a `stop` function:

{{< example file="examples/m03/iterators/iterators_test.go" region="pull" run="go test -run Example_pull -v ./examples/m03/iterators" >}}

Always call `stop` when you are done (a `defer` is the easy way), or the iterator's resources may never be released. Pull iterators are more expensive than ranging, so use them only when you need them.

## When to return an iterator

Return an iterator instead of a slice when:

- The sequence is **large or infinite**, and the caller might stop early.
- Producing each value is **expensive** (I/O, computation) and you want laziness.
- You want callers to **read but not modify** your internal data (an iterator over a slice field hides the slice).

Return a slice when the data is small, already in memory, and callers will want `len`, indexing or sorting.

## Common mistakes

### Ignoring the result of `yield`

An iterator that keeps calling `yield` after it returns `false` panics at run time. Always write `if !yield(v) { return }`.

### Doing work before the first `yield` that cannot be cancelled

If your iterator opens a file or a network connection, close it with `defer` inside the iterator function, so it is released when the loop ends early.

### Ranging over the same iterator twice and expecting the same results

Some iterators are single use (reading from a stream, for example). Document whether yours can be reused.

## Exercise

{{< exercise dir="m03/seq" >}}
Implement `Range`, `Map` and `Take`. The tests run your functions on an **infinite** iterator and count how many values were pulled, so `Take` must stop pulling as soon as it has enough, and every function must respect early `break`.
{{< /exercise >}}

## Key takeaways

- An iterator is a function that calls `yield` per value; `for range` can loop over it (Go 1.23+).
- `yield` returns `false` when the loop stops early, and the iterator must then return.
- Iterators compose into lazy pipelines; `slices.Collect` turns them back into slices.
- `slices`, `maps` and `strings` provide iterator functions; `iter.Pull` gives you `next` and `stop`.
- Return iterators for large, lazy or read-only sequences; return slices for small in-memory data.

## Further reading

- [Range over function types](https://go.dev/blog/range-functions) on the Go blog
- [Package iter](https://pkg.go.dev/iter)
- [For statements with range clause](https://go.dev/ref/spec#For_range) in the spec
