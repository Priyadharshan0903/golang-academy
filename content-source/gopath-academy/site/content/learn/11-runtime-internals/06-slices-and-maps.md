---
title: "11.6 Slices and maps under the hood"
linkTitle: "11.6 Slices and maps"
weight: 6
description: "How append grows slices (and why the exact steps change between Go versions), how Go 1.24's Swiss table maps work, and why maps never give memory back, with measurements."
goVersion: "1.27"
---

## What you will learn

- How `append` chooses a new capacity, and why you should not rely on the exact numbers
- How the compiler can keep small slices on the stack (Go 1.25 and 1.26)
- How Go's maps work since Go 1.24: **Swiss tables**
- Why deleting from a map does not free memory, and what to do about it

## How slices grow

A slice is a pointer, a length and a capacity (lessons 3.2 and 11.4). When `append` needs more room than the capacity, the runtime allocates a bigger backing array, copies the elements, and returns a slice pointing to the new array. The rule for the new capacity, from the runtime's `nextslicecap` function:

- If you append many elements at once and need more than double, you get what you need.
- Below 256 elements, the capacity **doubles**.
- From 256 up, it grows by `(cap + 768) / 4` each step: a smooth transition from 2x toward **1.25x** for large slices.
- The result is then **rounded up** to one of the memory allocator's size classes, so no allocated memory is wasted.

Growing geometrically makes `append` **amortized O(1)**: each element is copied only a constant number of times on average.

This code records every capacity a slice passes through:

{{< example file="examples/m11/growth/growth.go" region="caps" >}}

{{< example file="examples/m11/growth/growth_test.go" region="test" run="go test -v -run TestCapSteps ./examples/m11/growth" >}}

Here is the interesting part. The same test on the same machine with two Go versions:

```text
Go 1.24   heap slice:  [1 2 4 8 16 32 64 128 256 512]
          local slice: [1 2 4 8 16 32 64 128 256 512]

Go 1.27   heap slice:  [1 2 3 4 8 16 32 64 128 256 512]
          local slice: [4 8 16 32 64 128 256 512]
```

and for large slices (both versions): `512 848 1280 1792 2560 3408 5120 7168 9216 ...`, where you can see the growth factor fall from 2 toward 1.25, and the rounding to size classes.

Two things changed between versions:

- In the **local** slice, which never leaves its function, Go 1.27 starts with capacity 4. Go 1.25 and 1.26 taught the compiler to give such slices a small backing store **on the stack** (here 32 bytes, 4 ints) for their first appends, so short slices never touch the heap.
- Even the heap slice gets a different small-capacity sequence, a result of allocator and runtime changes.

The lesson: **the growth sequence is an implementation detail.** Code must never depend on `cap` after `append`. If you know the final size, preallocate with `make([]T, 0, n)`, which is faster than any growth policy and does not change between versions.

## How maps work: Swiss tables

Go 1.24 replaced the map implementation with one based on **Swiss tables**, a design created at Google for C++ (in the Abseil library). It was one of the changes the Go 1.24 release notes credit for a 2 to 3% average reduction in CPU overhead across representative benchmarks, and its higher load factor lowers memory use. Map behavior did not change; only speed and memory did.

The core ideas:

- Slots are organized in **groups of 8**. Each group has an 8-byte **control word**, one byte per slot. A control byte says the slot is empty, deleted, or in use; for slots in use, it holds **7 bits of the key's hash** (called H2).
- A lookup hashes the key. The other 57 bits (H1) choose the group where the search starts. Then it compares H2 with **all 8 control bytes at once** (with SIMD instructions on amd64, or a bit trick elsewhere). Only slots whose 7 bits match need a full key comparison, and a false match happens in about 1 slot in 128.
- If the group has no match but has an empty slot, the key is not in the map. Otherwise the search moves to the next group in a probe sequence.
- A table is filled to at most **7/8** before it grows, which keeps lookups short while wasting little memory.

Go adds one twist to the original design. Growing a hash table means rehashing everything, which would make a single insert into a huge map pause for a long time. So a Go map is split into **tables of at most 1024 slots**, selected by the top bits of the hash through a directory ("extendible hashing"). When a table fills, only that table grows or splits in two, so no single insert does more than a bounded amount of work.

The old design (before Go 1.24) used buckets of 8 entries with overflow buckets chained on, and moved entries to a doubled bucket array incrementally. The Swiss table opt-out (`GOEXPERIMENT=noswissmap`) existed in Go 1.24 and has since been removed; Go 1.27 rejects it.

## Maps never shrink

{{< example file="examples/m11/growth/growth_test.go" region="shrink" run="go test -v -run TestMapsDoNotShrink ./examples/m11/growth" >}}

On Go 1.24 and Go 1.27 alike:

```text
full: 20 MiB, deleted every key: 20 MiB, after clear: 20 MiB, replaced the map: 0 MiB
```

Deleting every key, or calling `clear(m)`, empties the map but keeps all its tables, ready to be reused. That is usually what you want: a map that was big once will likely be big again. But a long-lived map that grew during a traffic spike, a cache, a registry of connections, keeps its peak size forever.

The fix is to **replace** the map: make a new map sized for the current contents (`make(map[K]V, len(old))`), copy the live entries into it with a loop, and drop the old one, for example when its length falls far below its peak. The old tables become garbage and the GC frees them.

The same applies to slices: `s = s[:0]` keeps the backing array. To release a big backing array, set the slice to `nil` or copy the elements you need into a new, smaller slice (`slices.Clip` only trims the capacity in the slice header; it does not free memory).

## Common mistakes

### Depending on capacity after append

The sequence changes between Go versions, as measured above. Preallocate when you know the size.

### Expecting delete or clear to free memory

They do not. Replace long-lived maps that shrank a lot.

### Assuming map iteration order

It is deliberately randomized and was never specified. Sort the keys (lesson 3.4) when order matters.

## Key takeaways

- `append` doubles small slices and grows large ones toward 1.25x, rounded to allocator size classes; the exact steps are not part of the language and do change.
- Since Go 1.25 and 1.26, small slices that do not escape can start on the stack.
- Go 1.24 maps are Swiss tables: groups of 8 slots with a control word of 7-bit hash fragments, a 7/8 load limit, and tables of at most 1024 slots that grow independently.
- Maps and slices never give memory back on their own; replace them to release it.

## Further reading

- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable) by Michael Pratt, on the Go blog
- [Go Slices: usage and internals](https://go.dev/blog/slices-intro)
- [Arrays, slices (and strings): The mechanics of 'append'](https://go.dev/blog/slices)
