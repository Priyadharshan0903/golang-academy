---
title: "3.5 Maps: usage, iteration order, concurrency warning"
linkTitle: "3.5 Maps"
weight: 5
description: "Creating and using maps, the comma-ok idiom, sets and grouping, nil maps, struct values, random iteration order and why maps are not safe for concurrent writes."
goVersion: "1.27"
---

## What you will learn

- How to create, read, update and delete map entries
- The **comma-ok** idiom to tell "missing" from "zero"
- Idiomatic sets, counters and group-by with maps
- Why writing to a `nil` map panics
- Why you cannot assign to a struct field inside a map
- Iteration order and the concurrency rule

## The idea

A map is Go's built-in hash table: `map[K]V` maps keys of type `K` to values of type `V`. Keys must be **comparable** (usable with `==`): numbers, strings, booleans, pointers, channels, arrays and structs of comparable fields. Slices, maps and functions cannot be keys.

Like slices, a map variable holds a small reference to the real data, so passing a map to a function lets the function modify it.

## The basics

{{< example file="examples/m03/maps/maps_test.go" region="basics" run="go test -run Example_basics -v ./examples/m03/maps" >}}

Reading a missing key returns the **zero value** for `V`, with no error and no panic. That is convenient, but it means `ages["nobody"]` and a real age of `0` look the same. Use the **comma-ok** form `v, ok := m[k]` whenever the difference matters.

`delete(m, k)` on a missing key is a no-op, so you never need to check first.

## `nil` maps

{{< example file="examples/m03/maps/maps_test.go" region="nil" run="go test -run Example_nilMap -v ./examples/m03/maps" >}}

A `nil` map behaves like an empty map for **reading**, `len`, `range` and `delete`. **Writing** panics. This is the one place where Go's zero value is not ready to use, so always create maps before adding entries:

```go
m := make(map[string]int)          // empty map
m := make(map[string]int, 1000)    // with a size hint, avoids rehashing while it grows
m := map[string]int{"a": 1}        // literal
```

A struct with a map field needs that field initialized too, usually in a constructor function (Module 4).

## Idioms you will use constantly

### Counting

{{< example file="examples/m03/maps/maps_test.go" region="count" run="go test -run Example_counting -v ./examples/m03/maps" >}}

`counts[w]++` works on a missing key because the read returns `0` first. The zero value doing useful work again.

### Sets

Go has no set type. A map with empty-struct values is the idiom:

{{< example file="examples/m03/maps/maps_test.go" region="set" run="go test -run Example_set -v ./examples/m03/maps" >}}

`struct{}` takes zero bytes. `map[string]bool` also works and reads a little more simply (`if seen[x]`), at the cost of one byte per entry. Both are common.

### Grouping

{{< example file="examples/m03/maps/maps_test.go" region="group" run="go test -run Example_groupBy -v ./examples/m03/maps" >}}

Appending to a missing key works because the zero value of a slice is `nil`, and `append` to `nil` is fine.

## Struct values cannot be modified in place

{{< example file="examples/m03/maps/maps_test.go" region="structvalues" run="go test -run Example_structValues -v ./examples/m03/maps" >}}

`m[k].field = v` does not compile: `cannot assign to struct field accounts["ana"].balance in map`. Map elements are not addressable, because the map may move them in memory as it grows. Either read, modify and write back the whole value, or store pointers (`map[string]*account`) when you update fields often.

## Iteration order is random

`for k, v := range m` visits every entry, but **in an unspecified order that changes from run to run**. The Go runtime randomizes it deliberately, so that nobody can accidentally depend on it.

When you need a stable order (output, tests, hashing), sort the keys:

```go
for _, k := range slices.Sorted(maps.Keys(m)) {
	fmt.Println(k, m[k])
}
```

It is safe to `delete` entries from a map while ranging over it. Entries added during iteration may or may not be visited.

## Maps and concurrency

**Maps are not safe for concurrent use.** If one goroutine writes to a map while another reads or writes it, the program can crash with:

```text
fatal error: concurrent map writes
```

This is a fatal error, not a panic: it cannot be recovered. Protect shared maps with a `sync.Mutex` or `sync.RWMutex`, or in some specific cases use `sync.Map`. Module 6 covers both. Concurrent **reads** with no writer are fine.

## Common mistakes

### Writing to a map that was never created

The most common map panic. Look for struct fields of map type that a constructor forgot to `make`.

### Using the zero value to mean "missing"

`if m[k] == 0` cannot tell "absent" from "present with value 0". Use comma-ok.

### Depending on iteration order

Tests that compare printed map output are flaky. Sort keys or compare with `maps.Equal`.

### Sharing a map between goroutines without a lock

Works in testing, crashes under load. Run your tests with `-race` (Module 6) to catch it early.

### Expecting a deep copy from assignment

`b := a` for maps makes both variables refer to the **same** map. Use `maps.Clone(a)` for a shallow copy.

## Exercise

{{< exercise dir="m03/inventory" >}}
Implement `Add`, `Remove`, `Merge` and `LowStock`. `Remove` must not create entries for missing items and must delete items that reach zero. `Merge` must return a new map without modifying its inputs, and handle `nil` arguments. `LowStock` must return a sorted result.
{{< /exercise >}}

## Key takeaways

- `map[K]V` needs comparable keys; reading a missing key returns the zero value.
- Use `v, ok := m[k]` to distinguish missing from zero.
- A `nil` map is readable but panics on write; create maps with `make` or a literal.
- `map[T]struct{}` is a set; `m[k]++` counts; `m[k] = append(m[k], v)` groups.
- Map elements are not addressable: write back whole struct values or store pointers.
- Iteration order is random, and concurrent writes are a fatal error.

## Further reading

- [Go maps in action](https://go.dev/blog/maps) on the Go blog
- [Map types](https://go.dev/ref/spec#Map_types) and [Index expressions](https://go.dev/ref/spec#Index_expressions) in the spec
- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable), about the map implementation introduced in Go 1.24
