---
title: "11.5 Interfaces and method dispatch under the hood"
linkTitle: "11.5 Interfaces"
weight: 5
description: "What an interface value is in memory, how itabs make method calls and type assertions work, why a typed nil is not nil, and what interface and generic calls cost compared with direct calls, measured."
goVersion: "1.27"
---

## What you will learn

- The two-word layout of interface values, and what each word holds
- How **itabs** implement method calls and type assertions
- Why an interface holding a nil pointer is not nil, at the machine level
- The measured cost of interface calls and generic calls versus direct calls
- When that cost matters, and when it does not

## Two words

An interface value is two machine words, 16 bytes on 64-bit systems (lesson 11.4):

```text
any (empty interface)            Shape (interface with methods)
┌──────────────┐                 ┌──────────────┐
│ type  ───────┼──► type info    │ itab ────────┼──► itab: interface type, concrete type,
├──────────────┤                 ├──────────────┤          method table [Area → Square.Area]
│ data  ───────┼──► the value    │ data ────────┼──► the value
└──────────────┘                 └──────────────┘
```

- An **empty interface** (`any`) holds a pointer to the dynamic type's runtime type information, and a data pointer.
- A **non-empty interface** holds a pointer to an **itab** ("interface table") instead. The itab pairs one interface type with one concrete type, and contains the concrete type's method addresses in the order the interface lists them.

The **data word** points to the value. If the value is itself a pointer (or pointer-shaped, such as a map or channel), the pointer is stored directly. Other values are stored somewhere else and the data word points to them; that "somewhere" is a heap allocation unless the compiler can avoid it (the cases you measured in lesson 10.4: small integers, constants, zero-sized values and non-escaping interfaces).

You can peek at the two words with `unsafe`. This is for learning only: the layout is an internal detail.

{{< example file="examples/m11/iface/iface.go" region="words" >}}

{{< example file="examples/m11/iface/iface_test.go" region="example" run="go test -run Example -v ./examples/m11/iface" >}}

Two `*int` values in interfaces share the same type word, a string has a different one, and for a pointer the data word is the pointer itself.

## How method calls work

{{< example file="examples/m11/iface/iface.go" region="types" >}}

When you call `s.Area()` on a `Shape`, the compiled code:

1. loads the itab pointer from the interface's first word,
2. loads the address of `Area` from the itab's method table,
3. makes an **indirect call** to that address, passing the data word as the receiver.

Itabs are computed the first time a given (interface, concrete type) pair is needed, and cached in a global table. For conversions the compiler can see at compile time, such as `var s Shape = Square{}`, the itab is built statically and there is no lookup at all.

**Type assertions** use the same information:

- `v.(Square)`, asserting to a concrete type: compare the type or itab word with `Square`'s. One comparison: very cheap.
- `v.(Stringer)`, asserting to an interface: find or build the itab for `(Stringer, dynamic type)`, which needs a cache lookup. Still fast, but more work.
- A **type switch** is a sequence of these checks, optimized by the compiler.

## Why a typed nil is not nil

The `Example_typedNil` output above, `false true`, comes straight from the layout. An interface is nil only when **both** words are nil. Storing a nil `*Square` sets the type (or itab) word to `*Square` and the data word to nil, so the interface is not nil, even though calling a method on it may then dereference a nil pointer. This is the mechanism behind the nil interface trap from lesson 4.4, and the "nil error that is not nil" bug from Module 5: a function returning a nil `*MyError` as an `error`.

## What dispatch costs

Three ways to sum the areas of 1024 shapes:

{{< example file="examples/m11/iface/iface.go" region="sum" >}}

{{< example file="examples/m11/iface/iface_test.go" region="bench" >}}

Results on the 2-core test VM with Go 1.27 (`benchstat -col /call`, 6 runs):

```text
           │   direct    │     generic      │    interface     │  generic-iface   │
           │   sec/op    │ sec/op  vs base  │ sec/op  vs base  │ sec/op  vs base  │
Dispatch-2   627.7n ± 4%   2688.0n  +328%     2683.5n  +328%     3400.5n  +442%
```

Per element, that is about **0.6 ns** for the direct call against about **2.6 ns** through the interface. The difference is not only the indirect call: the direct call is **inlined**, so the loop becomes a few multiply and add instructions the CPU can pipeline, while an indirect call can never be inlined.

The generic version is as slow as the interface version, which surprises many people. Go compiles generic functions by **GC shape stenciling**: one compiled copy per "shape" of type argument, plus a **dictionary** that describes the actual type. Calling a method on a type parameter goes through the dictionary, an indirect call much like an interface call. Instantiated with an interface type (`generic-iface`), it pays for the dictionary and then the itab.

How to read these numbers:

- **2 ns per call does not matter** for code that does real work per call: an HTTP handler, a database query, a log line. Use interfaces freely for design: testing seams, plug-in points, the standard library's `io.Reader` and `io.Writer`.
- In **tight inner loops over many small items**, prefer concrete types, or check the profile. Generics are a tool for writing code once, not a performance optimization.
- **PGO** (lesson 10.5) helps exactly here: when the profile shows one concrete type dominates a call site, the compiler adds a direct, inlinable call for it (devirtualization).

## Common mistakes

### Returning a typed nil as an interface

Return a literal `nil` for "no error", not a nil pointer of a concrete type.

### Choosing generics for speed

They are about as fast as interfaces for method calls. Measure if it matters.

### Avoiding interfaces everywhere "for performance"

Nanoseconds per call are irrelevant outside hot loops, and interfaces make code testable.

## Key takeaways

- Interfaces are two words: type (or itab) and data. Non-pointer values are stored indirectly, sometimes with an allocation.
- An itab holds the method table for one (interface, concrete type) pair; method calls are indirect calls through it and cannot be inlined.
- An interface is nil only if both words are nil.
- Measured: about 0.6 ns per direct, inlined call against about 2.6 ns per interface or generic method call.
- Use interfaces for design; optimize only hot loops, with concrete types or PGO.

## Further reading

- [Go Data Structures: Interfaces](https://research.swtch.com/interfaces) by Russ Cox (from 2009; the itab idea is unchanged)
- [Generics implementation: GC shape stenciling](https://github.com/golang/proposal/blob/master/design/generics-implementation-gcshape.md)
