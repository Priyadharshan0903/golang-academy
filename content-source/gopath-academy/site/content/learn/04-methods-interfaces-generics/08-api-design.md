---
title: "4.8 Designing APIs: accept interfaces, return structs"
linkTitle: "4.8 API design"
weight: 8
description: "Put methods, interfaces and generics together into clean package APIs: consumer-defined interfaces, constructors, useful zero values, config structs and functional options."
goVersion: "1.27"
---

## What you will learn

- The rule "accept interfaces, return structs", and why it keeps packages decoupled
- When to write a constructor, and how to name it
- How to make zero values useful
- Two ways to handle optional settings: config structs and functional options
- A short checklist for reviewing a package API

## Accept interfaces, return structs

{{< example file="examples/m04/apidesign/apidesign_test.go" region="consumer" run="go test -run Example_acceptInterfacesReturnStructs -v ./examples/m04/apidesign" >}}

Two halves, each with a reason:

- **Accept interfaces.** `greeting` asks only for something with `GetUser`. It works with the in-memory store, a Postgres store, a cache, or a test fake, and its signature documents exactly what it depends on.
- **Return structs.** `NewMemoryStore` returns `*MemoryStore`, not an interface. Callers get every method (including `AddUser`), can see the real type in their editor, and can define their own narrow interfaces as needed. Returning an interface would hide methods and force callers to type-assert.

The interface `userGetter` is **unexported and lives with the consumer**. That is normal in Go. Interfaces describe what a piece of code needs, so they belong to that code.

There are sensible exceptions: return an interface when the concrete type is genuinely private and may vary (`errors.New` returns `error`; `io.MultiReader` returns `io.Reader`), or when the standard API of your domain is an interface (`hash.Hash`).

## Constructors

Go has no constructors in the language. By convention, a function named `New` or `NewThing` builds a ready-to-use value:

- `NewMemoryStore() *MemoryStore` in a package with several types.
- `New() *Client` when the package has one main type, so callers write `redis.New()` rather than `redis.NewRedis()` (no stutter, lesson 1.5).

Write a constructor when the zero value is not usable (a map must be created, a required dependency must be passed) or when invariants must be checked. If the constructor can fail, return `(*T, error)`.

## Make the zero value useful

When you can, design types so that `var x T` already works: `sync.Mutex`, `bytes.Buffer`, `strings.Builder` and `http.Client` all do. Techniques:

- Treat zero fields as "use the default" (see the config example below).
- Create maps lazily on first write instead of in a constructor.
- Avoid required fields that cannot have a sensible zero value.

A useful zero value means fewer constructors, fewer nil panics, and types that can be embedded or declared in structs without ceremony.

## Optional settings

Go has no default arguments or keyword arguments. Two patterns fill the gap.

### Config structs

{{< example file="examples/m04/apidesign/apidesign_test.go" region="config" run="go test -run Example_configStruct -v ./examples/m04/apidesign" >}}

Simple, explicit and easy to document. Every setting is a visible field, and zero means default. This is a good default choice, especially for internal packages and for settings loaded from config files.

### Functional options

{{< example file="examples/m04/apidesign/apidesign_test.go" region="options" run="go test -run Example_functionalOptions -v ./examples/m04/apidesign" >}}

Each option is a function that modifies the value being built. This pattern shines when:

- the zero value of a setting is meaningful and cannot double as "default" (a timeout of 0 meaning "no timeout");
- options need validation or side effects;
- you want to add options over time without changing the constructor signature, as many public libraries (gRPC, Zap, OpenTelemetry) do.

The cost is more code and slightly less discoverable documentation. Do not use functional options for two or three settings in an internal package; a struct is clearer.

## An API review checklist

Before publishing a package, check:

1. Does each exported name need to be exported?
2. Do functions accept the **smallest** interface they need?
3. Do constructors return concrete types?
4. Is the zero value of each exported type either useful or clearly documented as not usable?
5. Are there names that stutter (`user.UserService`)?
6. Are errors returned as `error`, never as concrete pointer types (lesson 4.4)?
7. Is it documented whether types are safe for concurrent use?
8. Would a caller need `any` or type assertions to use it? If so, consider generics or a better interface.

## Common mistakes

### Exporting an interface plus a single implementation

`type Service interface {...}` with `type service struct` and `func New() Service`. Callers lose access to the concrete type, and adding a method to the interface breaks every other implementation. Export the struct.

### Constructors that do work

Opening network connections or starting goroutines inside `New` makes types hard to test and use. Prefer a separate `Start` or `Connect` method, or accept already connected dependencies.

### Too many functional options for a small type

A `NewPoint(WithX(1), WithY(2))` API is ceremony. Use a struct literal.

## Key takeaways

- Accept small interfaces defined by the consumer; return concrete types from constructors.
- Name constructors `New` or `NewThing`, and return `(T, error)` when construction can fail.
- Design for a useful zero value.
- Use config structs by default; reach for functional options in public APIs with many evolving settings.

## Further reading

- [Go Code Review Comments: Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)
- [Effective Go: Constructors and composite literals](https://go.dev/doc/effective_go#composite_literals)
- [Self-referential functions and the design of options](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html) by Rob Pike
- [Dave Cheney: Functional options for friendly APIs](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
