---
title: "2.8 Pointers without fear"
linkTitle: "2.8 Pointers"
weight: 8
description: "What pointers are in Go, how pass-by-value works, why returning a pointer to a local variable is safe, and when you should and should not use pointers."
goVersion: "1.27"
---

## What you will learn

- The two pointer operators, `&` and `*`
- Why every Go function call copies its arguments, and how pointers change that
- How pointers to structs work, including automatic dereferencing
- Why returning a pointer to a local variable is safe in Go
- Clear rules for when to use a pointer and when not to

## The idea

A pointer holds the **memory address** of a value. If you come from Java, Python or JavaScript, you already use pointers all the time; those languages just hide them behind "references". Go makes them explicit, and in exchange gives you precise control over whether a function gets a copy or shares the original.

Go pointers are much safer than C pointers:

- **No pointer arithmetic.** You cannot add to an address or walk through memory (outside the `unsafe` package).
- **Garbage collected.** Memory stays alive as long as anything points to it. No dangling pointers, no manual `free`.
- **Nil is checked.** Dereferencing `nil` panics with a clear message instead of silently corrupting memory.

## `&` and `*`

- `&x` gives the address of `x`. If `x` is an `int`, `&x` has type `*int`, read as "pointer to int".
- `*p` gives the value that `p` points to. It works for reading and for writing.

{{< example file="examples/m02/pointers/pointers_test.go" region="basics" run="go test -run Example_basics -v ./examples/m02/pointers" >}}

The `*` symbol means two related things: in a **type** (`*int`) it means "pointer to", and in an **expression** (`*p`) it means "the value at".

## Go always passes by value

Every function argument is **copied** into the function. An `int` copies the number. A struct copies every field. A pointer copies the **address**, so both copies point to the same value:

{{< example file="examples/m02/pointers/pointers_test.go" region="passing" run="go test -run Example_passByValue -v ./examples/m02/pointers" >}}

There is no pass-by-reference in Go. When you want a function to change something you own, pass a pointer to it.

## Pointers to structs

Pointers are most common with structs. Go dereferences them for you when you access fields, so you write `p.Score` instead of `(*p).Score`:

{{< example file="examples/m02/pointers/pointers_test.go" region="structs" run="go test -run Example_structPointers -v ./examples/m02/pointers" >}}

Structs are covered in Module 3, and methods with pointer receivers (where this really pays off) in Module 4.

## Returning a pointer to a local variable is safe

In C, returning the address of a local variable is a serious bug, because the stack frame disappears. In Go it is normal and safe:

{{< example file="examples/m02/pointers/pointers_test.go" region="escape" run="go test -run Example_returnPointer -v ./examples/m02/pointers" >}}

The compiler performs **escape analysis**: if a variable's address outlives the function, the variable is allocated on the heap instead of the stack, and the garbage collector frees it when nothing points to it anymore. You never decide stack versus heap yourself. You can see the compiler's decisions with `go build -gcflags=-m`, which we explore in Module 10.

`new(T)` allocates a zero `T` and returns `*T`. In practice, `&T{...}` with a struct literal is far more common, because it lets you set fields at the same time.

## Nil pointers

The zero value of any pointer type is `nil`. Reading through a nil pointer panics:

{{< example file="examples/m02/pointers/pointers_test.go" region="nil" run="go test -run Example_nilPointer -v ./examples/m02/pointers" >}}

Guard with `if p == nil` wherever a nil pointer is a legitimate possibility, and document it. Where nil is never valid, let the panic happen; it points straight at the bug.

## When to use a pointer

Use a pointer when:

- **The function must modify the caller's value.** The most common reason.
- **The value is large** and copying it on every call would be wasteful. "Large" means many fields or big arrays, not a struct of three ints.
- **You need to express "no value".** A `*int` can be `nil`, while an `int` is always some number. This is common for optional fields in JSON or database rows.
- **The value must not be copied.** Types like `sync.Mutex` must stay in one place; pass a pointer to the struct containing them.

Do **not** reach for a pointer when:

- **The value is small and read-only**: ints, small structs, `time.Time`. Copying is cheap and avoids nil checks.
- **It is a slice, map, channel, function or interface.** These already contain internal pointers, so passing them by value is cheap, and a map passed by value can already be modified by the callee. (Slices have a subtlety around `append`, covered in lesson 3.2.)
- **Only to "be faster".** Pointers can make code slower by pushing values to the heap and adding GC work. Measure first.

## Common mistakes

### Dereferencing without a nil check where nil is possible

Values from optional JSON fields, lookups that can fail, or zero-value structs often hold `nil` pointers. Check before dereferencing, or make the API return a value and an `ok` flag instead.

### Pointers to slices and maps

`func add(m *map[string]int)` and `func f(s *[]int)` are almost never needed. Pass the map or slice directly; when a function needs to grow a slice, return the new slice (`s = append(s, x)` style).

### Taking the address of a range variable to keep it

Since Go 1.22, each iteration has its own variable, so `&v` in a loop gives a different pointer each time. That is correct, but it points to a **copy** of the element, not the element in the slice. Use `&items[i]` when you want to point into the slice.

### Using pointers everywhere "for performance"

A small struct passed by value often stays on the stack and is faster than a pointer to a heap value. Keep values by default and switch to pointers for the reasons listed above.

## Exercise

{{< exercise dir="m02/ptr" >}}
Implement `Swap`, `Clamp` (which must not panic on a nil pointer), `ValueOr` and `IntPtr`. The tests check nil handling and that `IntPtr` gives a fresh pointer on every call.
{{< /exercise >}}

## Key takeaways

- `&x` takes an address, `*p` reads or writes through it, and `*T` is a pointer type.
- Go passes everything by value. Pass a pointer when the callee must modify your value.
- Field access through a struct pointer is automatic: `p.Field`.
- Returning a pointer to a local is safe thanks to escape analysis and the garbage collector.
- Use pointers to mutate, to avoid copying large values, to represent "absent", or for types that must not be copied. Otherwise, prefer values.

## Further reading

- [Pointer types](https://go.dev/ref/spec#Pointer_types) and [Address operators](https://go.dev/ref/spec#Address_operators) in the spec
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
- [Go FAQ: How do I know whether a variable is allocated on the heap or the stack?](https://go.dev/doc/faq#stack_or_heap)
