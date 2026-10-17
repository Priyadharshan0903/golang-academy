---
title: "4.3 The empty interface, any, type assertions and type switches"
linkTitle: "4.3 any and type switches"
weight: 3
description: "What any is, how to get a concrete value back out with a type assertion, how type switches work, and how to check for optional capabilities."
goVersion: "1.27"
---

## What you will learn

- What `any` (the empty interface) is and when it is appropriate
- How a **type assertion** extracts a concrete value, and its two forms
- How a **type switch** branches on the dynamic type
- How to assert to an **interface** to check for optional behavior

## The idea

`interface{}` has no methods, so **every type satisfies it**. Since Go 1.18 it has a shorter name: `any`. They are exactly the same type.

`any` is the escape hatch for values whose type is genuinely not known until run time: `fmt.Println(a ...any)`, JSON decoded into `map[string]any`, or a container that holds anything. The price is that the compiler can no longer check what you do with the value. You have to ask at run time, with type assertions and type switches.

Before Go had generics, `any` (as `interface{}`) was also used for generic containers. Today, prefer generics (lesson 4.6) when the type is the same throughout and only varies between uses.

## Type assertions

{{< example file="examples/m04/assertions/assertions_test.go" region="assert" run="go test -run Example_typeAssertion -v ./examples/m04/assertions" >}}

- `v.(string)` returns the string, or **panics** if `v` does not hold a string.
- `n, ok := v.(int)` never panics. If the assertion fails, `ok` is `false` and `n` is the zero value.

Use the single-value form only when a wrong type is a programming error you want to crash on. Otherwise, use comma-ok.

## Type switches

To branch on several possible types, use a type switch:

{{< example file="examples/m04/assertions/assertions_test.go" region="switch" run="go test -run Example_typeSwitch -v ./examples/m04/assertions" >}}

Details that matter:

- In a case with **one type**, `x` has that type (`len(x)` works in the `string` case).
- In a case with **several types** (`case int, int64:`) or in `default`, `x` keeps the interface type, because the compiler cannot know which one it is.
- `case nil` matches a nil interface.
- Cases can be **interfaces**, like `error` and `fmt.Stringer`. The first matching case wins, so order matters: put specific cases before general ones.

## Asserting to an interface: optional capabilities

A type assertion can target an interface, which asks "does this value **also** have these methods?". The standard library uses this to unlock faster paths:

{{< example file="examples/m04/assertions/assertions_test.go" region="capability" run="go test -run Example_capabilityCheck -v ./examples/m04/assertions" >}}

Real examples of the pattern:

- `io.Copy` checks whether the source implements `io.WriterTo` or the destination implements `io.ReaderFrom`, and lets them copy directly.
- `http.ResponseWriter` values may also implement `http.Flusher` for streaming responses.
- `fmt` checks for `fmt.Stringer`, `error` and `fmt.Formatter` before printing.

This keeps the required interface small (just `io.Reader`) while still letting capable types do better.

## When `any` is the wrong tool

Signs that `any` is hiding a design problem:

- A function takes `any` and immediately type-switches over three known types. Consider three functions, or an interface with a method each type implements.
- A struct field is `any` but always holds the same type. Use that type, or make the struct generic.
- `map[string]any` is passed deep into your program after JSON decoding. Decode into a struct with the fields you expect instead (lesson 8.2).

## Common mistakes

### Single-value assertions on untrusted data

`data["age"].(float64)` panics the moment the JSON has a string there. Use comma-ok and handle the failure.

### Expecting JSON numbers to be `int`

Decoding into `any` produces `float64` for every number. A type switch with only `case int:` never matches.

### Wrong case order in type switches

`case fmt.Stringer:` before `case MyType:` makes the second case unreachable if `MyType` has a `String` method.

### Asserting to the wrong pointer-ness

If an interface holds a `*User`, then `v.(User)` fails. The dynamic type must match exactly.

## Key takeaways

- `any` is `interface{}`: every type satisfies it, so the compiler cannot help you.
- `v.(T)` panics on mismatch; `v, ok := v.(T)` does not.
- A type switch branches on the dynamic type; single-type cases give you a typed variable.
- Asserting to an interface checks for optional capabilities, a pattern used throughout the standard library.
- Prefer concrete types, interfaces with methods, or generics over passing `any` around.

## Further reading

- [Type assertions](https://go.dev/ref/spec#Type_assertions) and [Type switches](https://go.dev/ref/spec#Type_switches) in the spec
- [Effective Go: Interface conversions and type assertions](https://go.dev/doc/effective_go#interface_conversions)
- [The Laws of Reflection](https://go.dev/blog/laws-of-reflection), which explains interface values in depth
