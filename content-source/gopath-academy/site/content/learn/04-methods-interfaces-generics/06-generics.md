---
title: "4.6 Generics: type parameters and constraints"
linkTitle: "4.6 Generics"
weight: 6
description: "Write functions and types that work across many types with type parameters, constrain them with any, comparable, cmp.Ordered and type unions, and rely on type inference."
goVersion: "1.27"
---

## What you will learn

- How to declare generic functions and generic types
- What a **constraint** is: `any`, `comparable`, `cmp.Ordered`, method constraints and type unions
- What `~` means in a constraint
- How type inference saves you from writing type arguments
- How to get the zero value of a type parameter

## The idea

Generics (added in Go 1.18) let you write code once for many types, **without** losing compile-time type checking. Before generics you had two options: copy the code for every type, or use `any` and type assertions. Generics remove that trade-off for code like containers, algorithms over slices and maps, and helpers like `Map` or `Filter`.

A generic function declares **type parameters** in square brackets before its normal parameters. Each type parameter has a **constraint** that says which types are allowed and what you can do with them.

## Generic functions

{{< example file="examples/m04/generics/generics_test.go" region="func" run="go test -run Example_genericFunction -v ./examples/m04/generics" >}}

`Map[T, U any]` has two type parameters. You almost never write type arguments yourself: the compiler **infers** `T` and `U` from the arguments. Explicit type arguments (`Map[int, int]`) are needed only when inference has nothing to work from, for example when a type parameter appears only in the result.

## Constraints

A constraint is an interface. It limits which types can be used **and** decides which operations are allowed inside the function.

{{< example file="examples/m04/generics/generics_test.go" region="constraints" run="go test -run Example_constraints -v ./examples/m04/generics" >}}

| Constraint | Allows | Operations you can use |
|---|---|---|
| `any` | every type | assign, pass, return; no operators |
| `comparable` | types usable with `==` | `==`, `!=`, map keys |
| `cmp.Ordered` | integers, floats, strings | `<`, `<=`, `>`, `>=`, plus `==` |
| an interface with methods | types having those methods | call those methods |
| a type union | the listed types | operators all of them support |

Choose the **loosest** constraint that lets the function do its job. `IndexOf` only needs `==`, so it uses `comparable`, and callers can use it with structs. `MaxOf` needs `>`, so it requires `cmp.Ordered`.

### Type unions and `~`

{{< example file="examples/m04/generics/generics_test.go" region="union" run="go test -run Example_typeUnions -v ./examples/m04/generics" >}}

- `int | int64 | float64` is a **type union**: T must be one of those types, and you can use operators they all support (here `+`).
- `~float64` means "**any type whose underlying type is** `float64`". Without the tilde, `Celsius` would not be allowed, even though it is a `float64` underneath.
- Interfaces containing unions can only be used as constraints, not as ordinary variable types.

The `cmp.Ordered` constraint in the standard library is built exactly this way, as a union of `~int`, `~int8`, ..., `~float64` and `~string`.

### Method constraints

{{< example file="examples/m04/generics/generics_test.go" region="methodconstraint" run="go test -run Example_methodConstraint -v ./examples/m04/generics" >}}

Notice how `JoinNames` calls `Map` with a function on `T`: generic functions compose.

## Generic types

Types can have type parameters too. Methods on a generic type use the type's parameters (written in the receiver as `Stack[T]`):

{{< example file="examples/m04/generics/generics_test.go" region="type" run="go test -run Example_genericType -v ./examples/m04/generics" >}}

`var zero T` is how you get the zero value of a type parameter. It is needed whenever a generic function must return "nothing", as in `Pop` on an empty stack.

You must instantiate a generic type with concrete type arguments before using it: `Stack[string]`, `Stack[int]`. A `Stack[string]` and a `Stack[int]` are different, unrelated types.

## How generics compile

Go compiles generic code using a mix of specialization and runtime dictionaries (called "GC shape stenciling"). You do not need the details, but two consequences are worth knowing:

- Generic code is fully type checked at compile time. There are no runtime type errors from generics.
- Generic code is usually as fast as hand-written code, but not always. In hot paths, measure (Module 10) before assuming.

## Common mistakes

### Constraints that are too strict

`func Contains[T cmp.Ordered]` when you only use `==`. Callers cannot use it with structs or pointers. Use `comparable`.

### Forgetting `~` for named types

`Sum[T int | float64]` rejects `Celsius` and `time.Duration`. Add `~` unless you deliberately want only the exact types.

### Making everything generic

A function used with exactly one type does not need type parameters. Lesson 4.7 covers when generics hurt more than they help.

### Expecting methods to add new type parameters (before Go 1.27)

Until Go 1.27, methods could only use the type parameters of their receiver type. Lesson 4.7 covers the new generic methods.

## Exercise

{{< exercise dir="m04/set" >}}
Build a generic `Set[T comparable]` with `New`, `Add`, `Has`, `Len` and `Union`, plus a function `Sorted[T cmp.Ordered]`. Think about why `Sorted` is a function rather than a method: the set only requires `comparable`, but sorting needs ordering. The tests use strings, ints and a struct type.
{{< /exercise >}}

## Key takeaways

- Type parameters go in square brackets; the compiler usually infers type arguments.
- Constraints are interfaces that restrict types and enable operations: `any`, `comparable`, `cmp.Ordered`, method sets and type unions.
- `~T` includes all types whose underlying type is `T`.
- Generic types like `Stack[T]` must be instantiated, and `var zero T` gives a zero value.
- Pick the loosest constraint that works.

## Further reading

- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [An Introduction To Generics](https://go.dev/blog/intro-generics) on the Go blog
- [Type parameter declarations](https://go.dev/ref/spec#Type_parameter_declarations) in the spec
- [Package cmp](https://pkg.go.dev/cmp)
