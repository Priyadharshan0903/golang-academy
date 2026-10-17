---
title: "4.2 Interfaces: implicit satisfaction and small interfaces"
linkTitle: "4.2 Interfaces"
weight: 2
description: "How Go interfaces are satisfied without declarations, why small interfaces are better, how method sets decide what satisfies what, and how interfaces compose."
goVersion: "1.27"
---

## What you will learn

- How an interface is defined and satisfied, with no `implements` keyword
- Why small interfaces lead to better designs, and how they make testing easy
- What an interface value contains
- How **method sets** decide whether `T` or only `*T` satisfies an interface
- How to compose interfaces and check satisfaction at compile time

## The idea

An interface is a **set of method signatures**. Any type that has all of those methods satisfies the interface, automatically. There is no declaration like `class Rect implements Shape`.

{{< example file="examples/m04/interfaces/interfaces_test.go" region="implicit" run="go test -run Example_implicitSatisfaction -v ./examples/m04/interfaces" >}}

This has a big consequence: **the code that uses an interface can define it, without the implementations knowing about it.** You can write an interface that a type from the standard library, or from a third party package, already satisfies. Producers and consumers stay decoupled.

## Small interfaces

The most useful interfaces in Go have one or two methods: `io.Reader`, `io.Writer`, `fmt.Stringer`, `error`, `http.Handler`. The Go proverb says it directly: *the bigger the interface, the weaker the abstraction.*

{{< example file="examples/m04/interfaces/interfaces_test.go" region="small" run="go test -run Example_smallInterfaces -v ./examples/m04/interfaces" >}}

Because `alertOnHighLoad` only asks for something that can `Notify`, a test can pass a tiny in-memory recorder instead of a real email sender. No mocking framework needed. This is the main way Go code is made testable.

By convention, one-method interfaces are named after the method plus `-er`: `Reader`, `Writer`, `Notifier`, `Closer`, `Stringer`.

## Interface values

An interface value holds two things: the **dynamic type** of the value stored in it, and the **value** itself.

{{< example file="examples/m04/interfaces/interfaces_test.go" region="values" run="go test -run Example_interfaceValues -v ./examples/m04/interfaces" >}}

When you call a method on an interface value, Go looks up the method for the dynamic type and calls it. An interface with no type and no value is `nil`, and calling a method on it panics. Lesson 4.4 looks at the tricky case where the type is set but the value is a nil pointer.

## Method sets: `T` vs `*T`

Which methods does a type have, for the purpose of satisfying an interface?

| Type | Method set |
|---|---|
| `T` (a value) | methods with **value** receivers |
| `*T` (a pointer) | methods with value receivers **and** pointer receivers |

So if any method in an interface is implemented with a pointer receiver, only a **pointer** satisfies that interface:

{{< example file="examples/m04/interfaces/interfaces_test.go" region="methodset" run="go test -run Example_methodSets -v ./examples/m04/interfaces" >}}

Why the asymmetry? A value stored in an interface is a copy that is not addressable, so Go could not take its address to call a pointer method. Methods that modify state need a pointer, and that requirement carries over to interfaces.

The line `var _ Resetter = (*Buffer)(nil)` is a **compile-time assertion**. It costs nothing at run time, and the build fails if `*Buffer` ever stops satisfying `Resetter`. Use it when a type is meant to implement an interface but nothing in the package would otherwise prove it.

## Composing interfaces

Interfaces can embed other interfaces, combining their method sets:

{{< example file="examples/m04/interfaces/interfaces_test.go" region="compose" run="go test -run Example_embeddedInterfaces -v ./examples/m04/interfaces" >}}

The standard library does this throughout `io`: `io.ReadWriter`, `io.ReadCloser`, `io.ReadWriteCloser`. Build big interfaces from small ones, and let functions ask for only the small piece they need.

## Where to define interfaces

Define an interface **in the package that uses it**, not next to the implementation. A storage package should export a concrete `*PostgresStore`. The service that needs storage declares the two or three methods it actually calls. Lesson 4.8 develops this into the rule "accept interfaces, return structs".

Do not create an interface until you have a reason: a second implementation, a test double, or a function that should accept several types. An interface with exactly one implementation and no tests using it is usually just indirection.

## Common mistakes

### Large "god" interfaces

A `UserRepository` interface with 20 methods is hard to implement, hard to fake and couples every caller to all 20. Split it, or let each consumer declare what it uses.

### Defining interfaces on the producer side

Exporting `type Store interface` next to `type store struct` (and returning the interface) prevents callers from using any method you did not put in the interface. Return the concrete type.

### Forgetting the pointer when satisfying an interface

`cannot use b (variable of struct type Buffer) as Resetter value in variable declaration: Buffer does not implement Resetter (method Reset has pointer receiver)`. Pass `&b`.

### Pointers to interfaces

`*io.Reader` is almost never what you want. An interface value already holds a pointer when the dynamic type is a pointer. Pass interfaces by value.

## Exercise

{{< exercise dir="m04/kv" >}}
Implement `MemStore`, then two **decorators**: `Prefixed`, which namespaces keys, and `Counting`, which records hits, misses and sets. Both wrap any `Store`, so the last test stacks them on top of each other. Decide which receivers need to be pointers.
{{< /exercise >}}

## Key takeaways

- A type satisfies an interface by having its methods; there is no `implements`.
- Keep interfaces small, often a single method, named with `-er`.
- An interface value holds a dynamic type and a value.
- `*T` has both value and pointer methods; `T` only value methods. Pointer-receiver methods mean only `*T` satisfies.
- Define interfaces where they are used, and check satisfaction with `var _ I = (*T)(nil)`.

## Further reading

- [Interface types](https://go.dev/ref/spec#Interface_types) and [Method sets](https://go.dev/ref/spec#Method_sets) in the spec
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Go Proverbs](https://go-proverbs.github.io/), including "The bigger the interface, the weaker the abstraction"
- [Go FAQ: Why doesn't Go have "implements" declarations?](https://go.dev/doc/faq#implements_interface)
