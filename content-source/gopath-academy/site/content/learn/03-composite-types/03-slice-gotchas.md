---
title: "3.3 Slice gotchas: aliasing, append surprises, memory retention"
linkTitle: "3.3 Slice gotchas"
weight: 3
description: "The slice bugs that bite experienced developers: two appends overwriting each other, appends lost in functions, and small slices keeping huge arrays alive."
goVersion: "1.27"
---

## What you will learn

- How two slices that share a backing array can overwrite each other through `append`
- How to prevent aliasing with `slices.Clone` and the full slice expression
- Why an `append` inside a function can be invisible to the caller
- How a tiny sub-slice can keep a huge array in memory
- Safe ways to filter and delete elements

## The idea

Every gotcha in this lesson comes from one fact from lesson 3.2: **slices share their backing array, and `append` writes into spare capacity when there is any.** Once you can picture the backing array, all of these bugs become easy to predict.

## Gotcha 1: two appends, one array

{{< example file="examples/m03/slicegotchas/gotchas_test.go" region="alias" run="go test -run Example_appendAliasing -v ./examples/m03/slicegotchas" >}}

`base` has length 3 and capacity 10. Both `append` calls fit into that spare capacity, so **both write into index 3 of the same array**. `a` and `b` end up as two headers looking at the same memory, and the second write wins.

This bug is nasty because it depends on capacity. If `base` had been created with `make([]int, 3)` (no spare capacity), each `append` would allocate a new array and the code would appear to work. A refactor that adds a capacity hint can silently break it.

### The fix: make the copy explicit

{{< example file="examples/m03/slicegotchas/gotchas_test.go" region="clone" run="go test -run Example_avoidAliasing -v ./examples/m03/slicegotchas" >}}

- `slices.Clone(s)` returns an independent copy.
- The **full slice expression** `s[low:high:max]` sets the capacity to `max - low`. With `base[:3:3]` the capacity equals the length, so the next `append` must allocate. This is the idiom to use when you hand out a slice and do not want the receiver's appends to scribble over your data.

## Gotcha 2: `append` inside a function

{{< example file="examples/m03/slicegotchas/gotchas_test.go" region="func" run="go test -run Example_appendInFunction -v ./examples/m03/slicegotchas" >}}

The function gets a **copy of the header**. `append` updates the copy's length (and maybe its pointer), and that copy disappears when the function returns. The caller's header still says length 0.

The idiomatic fix is to **return the new slice**, exactly like `append` itself does. Passing `*[]string` also works, but it is rarely the clearest choice.

Note that **changing existing elements** inside a function does work, because the elements live in the shared array:

{{< example file="examples/m03/slicegotchas/gotchas_test.go" region="modify" run="go test -run Example_modifyElements -v ./examples/m03/slicegotchas" >}}

So a function that takes a slice can modify its elements but not its length. Document which one your function does.

## Gotcha 3: small slice, huge memory

A sub-slice keeps the **entire** backing array alive, because the garbage collector cannot free part of an array.

```go
data, _ := os.ReadFile("huge.log")   // 500 MB
header := data[:64]                  // still references all 500 MB
cache[name] = header                 // the 500 MB can never be freed
```

Clone the part you want to keep:

{{< example file="examples/m03/slicegotchas/gotchas_test.go" region="retain" run="go test -run Example_memoryRetention -v ./examples/m03/slicegotchas" >}}

The same applies to strings: `s[:10]` of a huge string shares its memory. Use `strings.Clone` when you keep a small piece of a large string for a long time.

A related effect: if a slice holds **pointers**, removing elements by reslicing (`s = s[:len(s)-1]`) leaves the removed pointer in the backing array, so the object it points to stays reachable. `slices.Delete` and `slices.Compact` clear the vacated elements for you since Go 1.22.

## Filtering and deleting safely

Removing elements while iterating with indexes is error-prone. Two clean patterns:

{{< example file="examples/m03/slicegotchas/gotchas_test.go" region="delete" run="go test -run Example_deleteWhileFiltering -v ./examples/m03/slicegotchas" >}}

- **Filter in place** with `keep := s[:0]`: it reuses the same array, which is fast and allocation free. It overwrites the original contents, so only do it when you own the slice.
- **`slices.DeleteFunc`** does the same thing with a predicate and clears the leftover tail.

## A mental checklist

When you see a slice crossing a boundary (a function call, a struct field, a channel, a cache), ask:

1. **Will anyone append to it?** If both sides might, give each side its own copy, or cap the capacity with `s[:len(s):len(s)]`.
2. **Will anyone modify elements?** If the caller expects its data to stay unchanged, clone before modifying.
3. **Is it a small piece of something big that will be kept for a long time?** Clone it.

## Common mistakes

### Returning internal slices from a type

A method that returns `t.items` lets callers modify your internal state and append into your spare capacity. Return `slices.Clone(t.items)`, or an iterator (lesson 3.8).

### Storing a sub-slice of a reused buffer

`bufio.Scanner.Bytes()` returns a slice that is **overwritten** on the next call to `Scan`. If you store it, you will later see different data. Copy it, or use `Text()`, which returns a new string.

### Forgetting to use the result of `slices.Compact` or `slices.Delete`

These functions work in place and return the shorter slice. The original variable still has the old length (and zeroed tail elements). Always assign the result: `s = slices.Delete(s, i, i+1)`.

## Key takeaways

- Slices that share a backing array can overwrite each other through `append` when there is spare capacity.
- Use `slices.Clone` or the full slice expression `s[a:b:b]` to stop aliasing.
- A function can modify a slice's elements but not its length; return the new slice instead.
- Sub-slices keep the whole backing array alive. Clone small pieces of big data you keep.
- Filter with `s[:0]` or `slices.DeleteFunc`, and always use the returned slice.

## Further reading

- [Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
- [Full slice expressions](https://go.dev/ref/spec#Slice_expressions) in the spec
- [Robust generic functions on slices](https://go.dev/blog/generic-slice-functions) on the Go blog, about zeroing in `slices.Delete` and friends
