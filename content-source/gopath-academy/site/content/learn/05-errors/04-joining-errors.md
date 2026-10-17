---
title: "5.4 errors.Join and multi-errors"
linkTitle: "5.4 Joining errors"
weight: 4
description: "Return several errors at once with errors.Join and multiple %w verbs, and inspect them with errors.Is, errors.As and Unwrap() []error."
goVersion: "1.27"
---

## What you will learn

- When reporting several errors at once is better than stopping at the first
- How `errors.Join` combines errors, and how it formats them
- How to wrap several errors with multiple `%w` verbs
- How `errors.Is` and `errors.As` search joined errors
- How to iterate over the individual errors

## The idea

Most functions stop at the first error. Some should not:

- **Validation**: tell the user about every invalid field in one response, not one per submission.
- **Cleanup**: when closing five resources, try all five and report every failure.
- **Fan-out work**: several independent operations ran; some failed.

Since Go 1.20 the standard library supports errors that wrap **more than one** error.

## `errors.Join`

{{< example file="examples/m05/join/join_test.go" region="join" run="go test -run Example_join -v ./examples/m05/join" >}}

- `errors.Join(errs...)` returns an error wrapping all non-nil arguments.
- It returns **`nil`** if every argument is `nil` (or there are none), so you can always `return errors.Join(errs...)`.
- Its message is the messages of the joined errors, separated by newlines.
- `errors.Is` and `errors.As` find any of the joined errors.

## Multiple `%w` verbs

`fmt.Errorf` can wrap several errors at once:

{{< example file="examples/m05/join/join_test.go" region="multiw" run="go test -run Example_multipleW -v ./examples/m05/join" >}}

Use this when you want a custom one-line message rather than the newline-separated format of `Join`.

## Getting the individual errors

An error that wraps several errors implements `Unwrap() []error` (plural). `errors.Unwrap` only understands the single-error form, which is why it returned `nil` above. To list the errors, assert to the interface:

```go
if multi, ok := err.(interface{ Unwrap() []error }); ok {
	for _, e := range multi.Unwrap() {
		fmt.Println("-", e)
	}
}
```

You will do exactly this in the exercise to collect the names of failed form fields.

## Joining errors from cleanup

A pattern for functions that must release several resources:

```go
func (s *Server) Close() error {
	return errors.Join(
		s.listener.Close(),
		s.db.Close(),
		s.cache.Close(),
	)
}
```

Every `Close` runs (Go evaluates all arguments), nil results are dropped, and the caller learns about every failure.

## Common mistakes

### Returning a slice of errors

`func Validate() []error` forces every caller to check `len(errs) > 0` and does not compose with `error`-returning code. Return `errors.Join(errs...)`.

### Expecting one-line messages from `Join`

The newline-separated output can look odd in a single-line log. Format it yourself with multiple `%w`, or log the joined error as a structured field.

### Calling `errors.Unwrap` on a joined error

It returns `nil`. Use `errors.Is`, `errors.As`, or the `Unwrap() []error` interface.

## Exercise

{{< exercise dir="m05/form" >}}
Implement `FieldError`, a `Validate` function that reports every problem in a signup form with `errors.Join`, and a `Fields` helper that lists the failed fields by unwrapping the joined error. The tests check both the exact message and that `errors.Is` finds every sentinel.
{{< /exercise >}}

## Key takeaways

- Report several errors at once for validation, cleanup and fan-out work.
- `errors.Join` returns nil for no errors, joins messages with newlines, and works with `errors.Is` and `errors.As`.
- `fmt.Errorf` accepts several `%w` verbs.
- Errors wrapping several errors implement `Unwrap() []error`, which `errors.Unwrap` does not follow.

## Further reading

- [Go 1.20 release notes: wrapping multiple errors](https://go.dev/doc/go1.20#errors)
- [Package errors: Join](https://pkg.go.dev/errors#Join)
