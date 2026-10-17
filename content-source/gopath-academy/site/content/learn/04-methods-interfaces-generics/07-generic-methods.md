---
title: "4.7 Generic methods (Go 1.27) and when generics are the wrong tool"
linkTitle: "4.7 Generic methods"
weight: 7
description: "Methods with their own type parameters arrived in Go 1.27. Learn the syntax, the limits around interfaces, and the cases where generics make code worse."
goVersion: "1.27"
---

## What you will learn

- How to declare a method with its **own** type parameters (Go 1.27+)
- Why generic methods cannot implement interface methods
- When generics are the right tool, and when interfaces or plain code are better

{{< callout type="info" >}}
The examples in this lesson need **Go 1.27 or newer**. Their files start with `//go:build go1.27`, so older toolchains skip them.
{{< /callout >}}

## Generic methods

Before Go 1.27, a method could only use the type parameters of its receiver. You could write `func (s *Stack[T]) Push(v T)`, but not a `Map` method that converts a `List[T]` into a `List[U]`, because `U` would be a new type parameter. The workaround was a top-level function, `Map(list, f)`.

Go 1.27 allows a method to declare its own type parameters:

{{< example file="examples/m04/genericmethods/genericmethods_test.go" region="method" run="go test -run Example_genericMethod -v ./examples/m04/genericmethods" >}}

As with generic functions, the type arguments are usually inferred: `nums.Map(strconv.Itoa)` infers `U = string`. You can also supply them explicitly: `nums.Map[int](...)`.

Generic methods also work on types that are not generic at all:

{{< example file="examples/m04/genericmethods/genericmethods_test.go" region="nongeneric" run="go test -run Example_methodOnPlainType -v ./examples/m04/genericmethods" >}}

The standard library uses them too. The Go 1.27 release notes mention `math/rand/v2`'s `(*Rand).N`, which is generic over the integer type.

## The interface limitation

Interface methods still **cannot** have type parameters, and a generic method **cannot satisfy** a non-generic interface method, even if one instantiation would match:

```go
type Converter interface {
	Convert(func(int) string) string
}

type Box struct{ n int }

func (b Box) Convert[T any](f func(int) T) T { return f(b.n) }

var _ Converter = Box{}
```

```text
cannot use Box{} (value of struct type Box) as Converter value in variable declaration:
Box does not implement Converter (wrong type for method Convert)
	have Convert[T any](func(int) T) T
	want Convert(func(int) string) string
```

The reason is that interface method calls are resolved at run time, while generic code is instantiated at compile time. The compiler cannot know which instantiations a dynamic call might need. So:

- Use generic methods for **convenience APIs** on concrete types (fluent transformations, typed accessors).
- Use **interfaces** when you need run-time polymorphism.

## When generics are the wrong tool

Generics are a power tool. Signs you are reaching for them too early:

**You only have one type.** A function only ever called with `[]User` should take `[]User`. Make it generic when the second type actually appears.

**An interface describes the behavior better.** If the function calls methods on its argument and nothing else, a plain interface parameter is simpler:

```go
// Generic, but no benefit:
func Describe[T fmt.Stringer](v T) string { return v.String() }

// Simpler, works the same for callers:
func Describe(v fmt.Stringer) string { return v.String() }
```

Generics win over interfaces when you need to **keep the concrete type** in the result (a `Max` that returns the same type it was given, a container that hands back `T` without type assertions), or when you need **operators** like `<` or `+`.

**The type parameter list is getting long.** Three or four type parameters with interdependent constraints make APIs hard to read and errors hard to understand. Often the design wants to be split up.

**You are simulating inheritance.** Generic "base" types with hooks are usually a sign that small interfaces and composition (lessons 3.7 and 4.2) would be cleaner.

A good rule from the Go team: **write the code without generics first. If you find yourself writing the same code a second time for a different type, then consider type parameters.**

## Common mistakes

### Using generic methods where callers need an interface

If several types must be handled through one interface variable, generic methods cannot help. Use interface methods with concrete types.

### Requiring Go 1.27 without saying so

A library with generic methods forces every user onto Go 1.27 or newer. Set the `go` line in `go.mod` accordingly, and mention it in your README.

### Generic code with constraints that leak implementation details

A constraint listing six specific types because the current implementation uses `+` is fragile. Prefer standard constraints like `cmp.Ordered` or small method sets.

## Key takeaways

- Since Go 1.27, methods can declare their own type parameters, on generic and non-generic types.
- Generic methods cannot implement interface methods; interfaces still cannot have type parameters.
- Use generics to preserve concrete types or use operators across types; use interfaces for behavior and run-time polymorphism.
- Write it concretely first, and generalize when the duplication is real.

## Further reading

- [Go 1.27 release notes: generic methods](https://go.dev/doc/go1.27)
- [When To Use Generics](https://go.dev/blog/when-generics) on the Go blog
- [Generic interfaces](https://go.dev/blog/generic-interfaces) on the Go blog
