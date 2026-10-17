---
title: "5.2 Wrapping with %w, errors.Is and errors.As"
linkTitle: "5.2 Wrapping errors"
weight: 2
description: "Add context to errors without losing the original, inspect wrapped errors with errors.Is, errors.As and errors.AsType, and make your own types unwrappable."
goVersion: "1.27"
---

## What you will learn

- How `%w` wraps an error and builds an error chain
- The difference between `%w` and `%v` in `fmt.Errorf`
- How `errors.Is` and `errors.As` search the chain
- The generic `errors.AsType` (Go 1.26+)
- How to make your own error types participate with `Unwrap`

## The idea

Each layer of a program should add context to an error ("load profile", "find user 42") so the final message explains what happened. But callers also need to recognize the **original** cause ("not found") to decide what to do.

Wrapping gives you both. `fmt.Errorf("context: %w", err)` produces a new error whose message includes the context **and** which remembers `err` underneath. The result is a chain you can search.

## Wrapping and `errors.Is`

{{< example file="examples/m05/wrapping/wrapping_test.go" region="wrap" run="go test -run Example_wrapping -v ./examples/m05/wrapping" >}}

After two layers of wrapping, `err == ErrNotFound` is false: `err` is a wrapper, not the sentinel itself. `errors.Is(err, ErrNotFound)` walks down the chain, comparing each error, and finds it.

**Always use `errors.Is` instead of `==` for errors**, even if you think nothing wraps them today. Someone will add context later.

## `%w` vs `%v`

{{< example file="examples/m05/wrapping/wrapping_test.go" region="verbs" run="go test -run Example_wrapVsFormat -v ./examples/m05/wrapping" >}}

The messages are identical, but only `%w` keeps the original error inspectable. When to use which:

- **`%w`** when callers may reasonably want to check the cause. This is the common case.
- **`%v`** when you deliberately want to **hide** the underlying error, so it does not become part of your API. For example, a storage package might not want callers to depend on the specific database driver errors it happens to use today.

## `errors.As`: finding a type in the chain

{{< example file="examples/m05/wrapping/wrapping_test.go" region="as" run="go test -run Example_errorsAs -v ./examples/m05/wrapping" >}}

`errors.As(err, &target)` looks for the first error in the chain that can be assigned to `target`'s type, and if found, stores it in `target` and returns true. You pass a **pointer to a variable** of the type you are looking for, here a `*fs.PathError`, so the argument is a `**fs.PathError`. `go vet` checks that you get this right.

Notice also that `errors.Is(err, fs.ErrNotExist)` works on the wrapped `os.Open` error. The standard library's file errors are designed to be checked this way, on every operating system.

### `errors.AsType` (Go 1.26+)

Go 1.26 added a generic version that needs no target variable:

{{< example file="examples/m05/wrapping/astype_test.go" region="astype" run="go test -run Example_asType -v ./examples/m05/wrapping" >}}

It is type safe at compile time (you cannot pass the wrong kind of pointer) and faster than `errors.As`. Prefer it in modules whose `go` line is 1.26 or later.

## Your own wrapping types

A custom error type can wrap another error by implementing `Unwrap() error`:

{{< example file="examples/m05/wrapping/wrapping_test.go" region="custom" run="go test -run Example_customWrapper -v ./examples/m05/wrapping" >}}

With `Unwrap`, both `errors.Is(err, ErrTimeout)` and `errors.As(err, &qe)` see through the type. Without it, the chain stops at `QueryError`.

`errors.Unwrap(err)` returns one level down, or `nil` if `err` does not wrap anything. You rarely call it directly; `Is` and `As` do the walking.

## How the chain is searched

For each error in the chain, starting with `err` itself:

- `errors.Is` reports true if the error `==` target, **or** the error has a method `Is(error) bool` that returns true for target.
- `errors.As` succeeds if the error is assignable to the target type, **or** has a method `As(any) bool` that returns true.
- Then it moves on to `Unwrap()`. If an error has `Unwrap() []error` (lesson 5.4), the search goes depth-first through each wrapped error.

Custom `Is` methods let one error match a category; lesson 5.3 shows an example.

## Common mistakes

### Wrapping with `%v` and then calling `errors.Is`

`%v` flattens to a string. If callers must check the cause, use `%w`.

### Passing a non-pointer to `errors.As`

`errors.As(err, pathErr)` instead of `errors.As(err, &pathErr)` panics. `go vet` catches it.

### Wrapping the same context at every layer

`read config: read config: open config.yaml: ...` happens when both caller and callee add the same words. Each layer should add only what **it** knows: the function name, the ID it was working on.

### Wrapping errors from your dependencies into your public API without thinking

Everything you wrap with `%w` becomes something callers might check. Be intentional about it, especially in libraries.

## Exercise

{{< exercise dir="m05/config" >}}
Build a small config parser. `ParseLine` returns errors wrapping `ErrSyntax`. `Load` returns a `*ParseError` with the line number, whose `Unwrap` method exposes the cause, so callers can use both `errors.As` (to get the line) and `errors.Is` (to check `ErrSyntax` or `ErrDuplicateKey`).
{{< /exercise >}}

## Key takeaways

- `fmt.Errorf("context: %w", err)` adds context and keeps the original error in a chain.
- Compare errors with `errors.Is`, never `==`.
- Extract typed errors with `errors.As(err, &target)`, or `errors.AsType[T](err)` in Go 1.26+.
- Use `%v` to hide an underlying error on purpose.
- Implement `Unwrap() error` on wrapping error types.

## Further reading

- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) on the Go blog
- [Package errors](https://pkg.go.dev/errors)
- [Go 1.26 release notes](https://go.dev/doc/go1.26) for `errors.AsType`
