---
title: "3.1 Arrays and why you rarely use them directly"
linkTitle: "3.1 Arrays"
weight: 1
description: "Go arrays have a fixed length that is part of their type and are copied by value. Learn where that is useful and why slices are the everyday choice."
goVersion: "1.27"
---

## What you will learn

- How arrays are declared and why their length is part of the type
- Why assigning or passing an array copies all of it
- The few places arrays are the right tool
- How arrays relate to slices, which you will use instead most of the time

## The idea

A Go array is a **fixed-size sequence of elements of one type**. The size is part of the type: `[3]int` and `[4]int` are different, incompatible types.

Arrays are **values**, not references. Assigning an array to another variable, or passing it to a function, copies every element. This is the opposite of arrays in Java, Python or JavaScript, where the variable is a reference to shared data.

{{< example file="examples/m03/arrays/arrays_test.go" region="basics" run="go test -run Example_basics -v ./examples/m03/arrays" >}}

## Value semantics

{{< example file="examples/m03/arrays/arrays_test.go" region="value" run="go test -run Example_valueSemantics -v ./examples/m03/arrays" >}}

Copying is predictable and safe (no one else can change your copy), but for large arrays it is expensive. A function that takes a `[1_000_000]byte` copies a megabyte on every call.

Arrays whose element type is comparable can be compared with `==`, element by element.

## Where arrays are the right tool

You will rarely declare arrays in application code, but they do show up:

- **Fixed-size values with meaning**: a SHA-256 hash is `[32]byte` (`sha256.Sum256` returns exactly that), an IPv4 address fits in `[4]byte`, an RGB color in `[3]uint8`.
- **Map keys**: because arrays are comparable values, they work as keys where a slice cannot.
- **The backing store of a slice**: every slice points to an array underneath. That is the topic of the next lesson.

{{< example file="examples/m03/arrays/arrays_test.go" region="key" run="go test -run Example_asMapKey -v ./examples/m03/arrays" >}}

## Arrays vs slices at a glance

| | Array `[N]T` | Slice `[]T` |
|---|---|---|
| Length | Fixed, part of the type | Can grow with `append` |
| Assignment and passing | Copies all elements | Copies a small header; elements are shared |
| Comparable with `==` | Yes (if elements are) | No (only to `nil`) |
| Usable as a map key | Yes | No |
| Typical use | Hashes, fixed-size buffers, keys | Almost everything else |

## Common mistakes

### Expecting a function to modify an array argument

```go
func reset(a [3]int) { a[0] = 0 } // modifies a copy
```

Pass a pointer (`*[3]int`) or, much more commonly, use a slice.

### Confusing `[]int{1, 2}` and `[2]int{1, 2}`

The first is a slice literal, the second an array literal. A missing number or `...` between the brackets is the only difference, and they behave very differently.

### Ranging over a large array by value

`for i, v := range bigArray` evaluates `bigArray` once, which copies it. Range over a slice of it instead (`bigArray[:]`) or over a pointer to it.

## Key takeaways

- An array's length is part of its type, and it never changes.
- Arrays are values: assignment and function calls copy every element.
- Use arrays for fixed-size values like hashes and for comparable map keys.
- For general lists of data, use slices.

## Further reading

- [Array types](https://go.dev/ref/spec#Array_types) in the spec
- [Go Slices: usage and internals](https://go.dev/blog/slices-intro), which starts with arrays
