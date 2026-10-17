---
title: "3.2 Slices: length, capacity and the backing array"
linkTitle: "3.2 Slices"
weight: 2
description: "What a slice really is (a pointer, a length and a capacity), how slicing shares memory, how append grows, and the difference between nil and empty slices."
goVersion: "1.27"
---

## What you will learn

- What a slice is under the hood: pointer, length and capacity
- How slicing an array or slice creates a **window** onto shared memory
- How `append` decides whether to reuse memory or allocate
- The difference between a `nil` slice and an empty slice
- How to use `make`, `copy` and multi-dimensional slices

## The idea

A slice is a small descriptor with three fields:

```text
slice header
+---------+-----+-----+
| pointer | len | cap |
+----|----+-----+-----+
     |
     v
backing array:  [ a | b | c | d | e | f ]
```

- **pointer**: where the slice's first element lives in an underlying array.
- **len**: how many elements the slice currently contains.
- **cap**: how many elements fit before the end of the underlying array.

When you pass a slice to a function or assign it, you copy this header (24 bytes on 64-bit systems). The elements are **not** copied. Both headers point at the same array.

## Creating slices

{{< example file="examples/m03/slices/slices_test.go" region="basics" run="go test -run Example_basics -v ./examples/m03/slices" >}}

- A **literal** `[]int{10, 20, 30}` creates an array and a slice over all of it.
- `make([]T, len)` creates a slice of `len` zero values.
- `make([]T, len, cap)` also reserves room to grow without reallocating. Use it when you know roughly how many elements you will add.

## Slicing shares memory

`s[low:high]` creates a new header that points into the same array. The new slice contains elements `low` up to (but not including) `high`:

{{< example file="examples/m03/slices/slices_test.go" region="header" run="go test -run Example_sharedBackingArray -v ./examples/m03/slices" >}}

Two things to notice:

- `cap(s)` is 4, not 3. Capacity counts from the slice's start to the **end of the underlying array**, because the slice is allowed to grow into that space.
- Writing `s[0]` changed `arr[1]`. A slice is a window, not a copy.

Shortcuts: `s[:n]` is the first `n` elements, `s[n:]` drops the first `n`, and `s[:]` covers everything.

## How `append` works

`append(s, values...)` always returns a slice, and you must use the result:

1. If `len(s) + len(values) <= cap(s)`, it writes the new values into the existing backing array, right after the last element, and returns a header with a bigger length. **No allocation.**
2. Otherwise it allocates a **new, larger** array, copies the old elements over, adds the new ones and returns a header pointing to the new array.

{{< example file="examples/m03/slices/slices_test.go" region="append" run="go test -run Example_appendGrowth -v ./examples/m03/slices" >}}

When it grows, Go roughly doubles the capacity for small slices and grows by a smaller factor for large ones. The exact amounts are an implementation detail, so never write code that depends on them (the example checks `cap(s) >= 3` rather than an exact number for that reason).

Because growth may or may not allocate, **after an `append` you cannot know whether the result shares memory with the original**. This is the root of the gotchas in the next lesson.

{{< callout type="info" >}}
Always write `s = append(s, x)`. The compiler rejects an `append` whose result is not used, because forgetting to assign it is always a bug.
{{< /callout >}}

## `nil` vs empty slices

{{< example file="examples/m03/slices/slices_test.go" region="nil" run="go test -run Example_nilVsEmpty -v ./examples/m03/slices" >}}

A `nil` slice has no backing array, and it behaves exactly like an empty slice for `len`, `cap`, `range` and `append`. Prefer `var s []T` when starting an empty list.

The difference only shows when something looks at `nil` specifically. The most common case is JSON: a `nil` slice encodes as `null` and an empty slice as `[]`. If an API must return `[]`, initialize with `[]T{}` or `make([]T, 0)`.

## `copy`

`copy(dst, src)` copies elements between slices and returns how many it copied, which is the smaller of the two lengths:

{{< example file="examples/m03/slices/slices_test.go" region="copy" run="go test -run Example_copy -v ./examples/m03/slices" >}}

`copy` never allocates and never grows `dst`. To duplicate a slice, use `slices.Clone(s)` (lesson 3.4).

## Two-dimensional slices

Go has no built-in matrix type. A `[][]T` is a slice of slices, and each inner slice must be created separately:

{{< example file="examples/m03/slices/slices_test.go" region="grid" run="go test -run Example_twoDimensional -v ./examples/m03/slices" >}}

## Common mistakes

### Indexing past the length

`s := make([]int, 0, 10); s[0] = 1` panics with `index out of range [0] with length 0`. Capacity is room to grow, not elements you can index. Use `append`, or `make([]int, 10)` if you want ten indexable zeros.

### `make([]T, n)` followed by `append`

```go
ids := make([]int, len(users)) // already has len(users) zeros
for _, u := range users {
	ids = append(ids, u.ID) // appends AFTER the zeros
}
```

Use `make([]int, 0, len(users))` with `append`, or `make([]int, len(users))` with `ids[i] = ...`.

### Comparing slices with `==`

Slices are only comparable to `nil`. Use `slices.Equal(a, b)`.

### Assuming a sub-slice is a copy

Writing through a sub-slice changes the original. Clone it when you need independence.

## Exercise

{{< exercise dir="m03/ints" >}}
Implement `Chunk`, `RemoveAt` and `Dedupe`. The tests check that results never share memory with their inputs, so a solution that works on a quick glance but aliases the backing array will fail. Think about capacity.
{{< /exercise >}}

## Key takeaways

- A slice is a header of pointer, length and capacity over a backing array.
- Assigning, passing and slicing copy the header, not the elements.
- `append` reuses spare capacity when it can and reallocates when it cannot. Always use its result.
- `nil` and empty slices behave the same, except where code checks for `nil` (such as JSON encoding).
- Use `make` with a capacity hint when you know the size, and `copy` or `slices.Clone` to duplicate.

## Further reading

- [Go Slices: usage and internals](https://go.dev/blog/slices-intro) on the Go blog
- [Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
- [Slice types](https://go.dev/ref/spec#Slice_types) and [Appending to and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices) in the spec
