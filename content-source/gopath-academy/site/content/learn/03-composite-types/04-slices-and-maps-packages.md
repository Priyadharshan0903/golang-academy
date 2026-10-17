---
title: "3.4 The slices and maps packages"
linkTitle: "3.4 slices and maps packages"
weight: 4
description: "The generic helper packages for searching, sorting, editing and cloning slices and maps, and the built-in min, max and clear functions."
goVersion: "1.27"
---

## What you will learn

- The most useful functions in `slices` and `maps`
- How to sort with a custom comparison using `slices.SortFunc` and `cmp`
- Which functions modify their input and which return copies
- The built-ins `min`, `max` and `clear`

## The idea

For years, Go developers wrote the same small loops again and again: "does this slice contain x?", "sort these structs by name", "copy this map". Since Go 1.21 the standard library has generic packages for these jobs:

- **`slices`**: searching, sorting, comparing, inserting, deleting and cloning slices.
- **`maps`**: cloning, comparing, copying and iterating maps.
- **`cmp`**: comparison helpers used for sorting.

They work with any element type (they are generic, Module 4). Use them instead of hand-written loops: they are tested, fast, and make intent obvious.

## Searching, sorting and comparing

{{< example file="examples/m03/stdslices/stdslices_test.go" region="slices" run="go test -run Example_slicesPackage -v ./examples/m03/stdslices" >}}

Look closely at the `Compact` line. `Compact` works **in place**: it moves the unique elements to the front and returns a shorter slice. The original variable still has length 5, and since Go 1.22 its leftover tail is zeroed, so `sorted` now prints `[2 5 8 9 0]`. Always use the returned slice.

| Need | Function | Modifies input? |
|---|---|---|
| Find an element | `slices.Contains`, `Index`, `IndexFunc` | no |
| Smallest, largest | `slices.Min`, `Max` (panic on empty) | no |
| Sort | `slices.Sort`, `SortFunc`, `SortStableFunc` | **yes** |
| Binary search on sorted data | `slices.BinarySearch`, `BinarySearchFunc` | no |
| Compare | `slices.Equal`, `Compare` | no |
| Remove consecutive duplicates | `slices.Compact`, `CompactFunc` | **yes**, use the result |
| Copy | `slices.Clone` | no |
| Reverse | `slices.Reverse` | **yes** |

## Sorting by your own rules

`slices.SortFunc` takes a comparison function that returns a negative number, zero or a positive number. `cmp.Compare` produces exactly that, and `cmp.Or` returns its first non-zero argument, which makes multi-key sorts read naturally:

{{< example file="examples/m03/stdslices/stdslices_test.go" region="sortfunc" run="go test -run Example_sortFunc -v ./examples/m03/stdslices" >}}

Swap `a` and `b` in a `cmp.Compare` call to reverse the order for that key. Use `slices.SortStableFunc` when elements that compare equal must keep their original order.

{{< callout type="info" >}}
You will still see `sort.Slice` and `sort.Strings` in older code. They work, but `slices.Sort` and `slices.SortFunc` are faster and type safe. Prefer them in new code.
{{< /callout >}}

## Inserting and deleting

{{< example file="examples/m03/stdslices/stdslices_test.go" region="edit" run="go test -run Example_insertAndDelete -v ./examples/m03/stdslices" >}}

`Insert` and `Delete` return the updated slice. Like `append`, they may reuse the backing array, so the old variable should not be used afterwards. Assign the result back.

## The `maps` package

{{< example file="examples/m03/stdslices/stdslices_test.go" region="maps" run="go test -run Example_mapsPackage -v ./examples/m03/stdslices" >}}

- `maps.Keys(m)` and `maps.Values(m)` return **iterators** (lesson 3.8), not slices. `slices.Sorted(maps.Keys(m))` collects the keys into a sorted slice in one step, which is the standard way to iterate a map in a stable order.
- `maps.Clone` makes a shallow copy. The values are copied as values, so if they are pointers or slices, both maps still point at the same underlying data.
- `maps.Equal` compares two maps key by key.

## Built-ins: `min`, `max`, `clear`

Since Go 1.21 these are built into the language, no import needed:

- `min(a, b, ...)` and `max(a, b, ...)` work on numbers and strings.
- `clear(m)` deletes every key from a map. `clear(s)` sets every element of a slice to its zero value (it does **not** change the length).

## Common mistakes

### Sorting a slice you do not own

`slices.Sort(userSlice)` reorders the caller's data. Clone first if the caller does not expect that.

### Ignoring results of in-place functions

`slices.Compact(s)` without `s = ...` leaves you with a slice that has stale length and zeroed elements at the end.

### Calling `slices.Min` or `slices.Max` on an empty slice

They panic. Check `len(s) > 0` first.

### Assuming `maps.Clone` is a deep copy

A `map[string][]int` clone shares every inner slice with the original.

## Key takeaways

- Reach for `slices`, `maps` and `cmp` before writing loops by hand.
- Sort with `slices.Sort` or `slices.SortFunc` plus `cmp.Compare` and `cmp.Or`.
- `Sort`, `Reverse`, `Compact`, `Insert` and `Delete` modify in place; always use the returned slice.
- `slices.Sorted(maps.Keys(m))` gives you map keys in a stable order.
- `min`, `max` and `clear` are built-ins.

## Further reading

- [Package slices](https://pkg.go.dev/slices), [maps](https://pkg.go.dev/maps) and [cmp](https://pkg.go.dev/cmp)
- [Robust generic functions on slices](https://go.dev/blog/generic-slice-functions) on the Go blog
- [Built-in functions min, max and clear](https://go.dev/ref/spec#Min_and_max) in the spec
