---
title: "4.1 Methods: value vs pointer receivers"
linkTitle: "4.1 Methods"
weight: 1
description: "Declaring methods on any named type, choosing between value and pointer receivers, handling nil receivers, and using method values and expressions."
goVersion: "1.27"
---

## What you will learn

- How to declare a method, and that any named type can have methods
- The difference between **value** and **pointer** receivers, and how to choose
- How Go automatically takes addresses and dereferences for method calls
- Why a method on a nil pointer can be valid
- Method values and method expressions

## The idea

A method is a function with a **receiver**: a special parameter written before the function name. `func (r Rect) Area() float64` is a function called `Area` that is attached to `Rect` and called as `r.Area()`.

Go has no classes. Instead, **any named type declared in your package** can have methods: structs, but also slices, maps, numbers, strings or functions. The method and the type must be in the same package, so you cannot add methods to `int` or to a type from another package (define your own named type instead).

## Value and pointer receivers

{{< example file="examples/m04/methods/methods_test.go" region="basic" run="go test -run Example_receivers -v ./examples/m04/methods" >}}

- **Value receiver** `(r Rect)`: the method gets a **copy**. It can read fields, and any changes affect only the copy.
- **Pointer receiver** `(r *Rect)`: the method gets the address. It can modify the original, and nothing is copied.

Go handles the `&` and `*` for you on calls: `r.Scale(2)` on an addressable `Rect` variable becomes `(&r).Scale(2)`, and calling a value method through a pointer dereferences automatically.

### The classic bug

{{< example file="examples/m04/methods/methods_test.go" region="valuebug" run="go test -run Example_valueReceiverBug -v ./examples/m04/methods" >}}

`IncBroken` compiles, runs and does nothing visible. If a method needs to change its receiver, it must take a pointer.

### How to choose

Use a **pointer receiver** when:

- the method modifies the receiver;
- the struct is large, so copying it on every call would be wasteful;
- the struct contains something that must not be copied, such as a `sync.Mutex`;
- **any other method on the type already uses a pointer receiver**. Consistency matters: mixing receiver kinds on one type confuses readers and complicates which interfaces the type satisfies (lesson 4.2).

Use a **value receiver** when the type is small and naturally immutable: `time.Time`, a `Point`, a `Celsius`, a small `Money` struct. Values are cheap to copy and safe to share.

When in doubt, use a pointer receiver for structs.

## Methods on non-struct types

{{< example file="examples/m04/methods/methods_test.go" region="anytype" run="go test -run Example_methodsOnAnyType -v ./examples/m04/methods" >}}

This is common in the standard library: `time.Duration` is an `int64` with methods, `http.HandlerFunc` is a function type with a `ServeHTTP` method, and `sort.StringSlice` is a `[]string` with methods.

## Nil receivers

A pointer receiver can be `nil`, and calling the method is **not** an automatic panic. The method runs and can check for `nil` itself:

{{< example file="examples/m04/methods/methods_test.go" region="nilrecv" run="go test -run Example_nilReceiver -v ./examples/m04/methods" >}}

This makes recursive data structures elegant: the empty list or empty tree is just `nil`, and methods handle it. The panic only happens if the method **dereferences** the nil pointer (reading `n.Val` without a check).

## Method values and method expressions

Methods can be used as function values in two ways:

{{< example file="examples/m04/methods/methods_test.go" region="values" run="go test -run Example_methodValues -v ./examples/m04/methods" >}}

- A **method value** `r.Area` binds the method to a specific receiver. For a value receiver, the receiver is **copied at that moment**, which is why changing `r.W` afterwards did not change the result.
- A **method expression** `(*Rect).Scale` turns the method into a plain function whose first parameter is the receiver.

Method values are handy for callbacks: `http.HandleFunc("/", srv.handleIndex)`, `defer mu.Unlock()`, or passing `strings.ToUpper` to a mapping function.

## Common mistakes

### Value receivers on methods that mutate

The `IncBroken` bug. If you see a method assigning to a field of a value receiver, it is almost always wrong.

### Mixing receiver kinds on one type

Pick one. If any method needs a pointer receiver, make all of them pointer receivers.

### Copying a struct that has pointer-receiver methods and a mutex

Passing such a struct by value copies the mutex, and `go vet` reports `passes lock by value`. Pass pointers.

### Calling pointer methods on non-addressable values

`Rect{1, 2}.Scale(2)` does not compile, because a composite literal used directly is not addressable. The same goes for map elements: `m["k"].Scale(2)`. Assign to a variable first, or store pointers.

## Key takeaways

- A method is a function with a receiver, and any named type in your package can have methods.
- Value receivers get a copy; pointer receivers can modify the original.
- Use pointer receivers to mutate, for large or non-copyable structs, and for consistency.
- A nil pointer receiver is valid until the method dereferences it.
- Method values bind a receiver; method expressions make the receiver an explicit parameter.

## Further reading

- [Method declarations](https://go.dev/ref/spec#Method_declarations) and [Method values](https://go.dev/ref/spec#Method_values) in the spec
- [Go FAQ: Should I define methods on values or pointers?](https://go.dev/doc/faq#methods_on_values_or_pointers)
- [Effective Go: Pointers vs. Values](https://go.dev/doc/effective_go#pointers_vs_values)
