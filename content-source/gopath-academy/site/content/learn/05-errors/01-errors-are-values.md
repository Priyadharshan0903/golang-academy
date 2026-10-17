---
title: "5.1 Errors are values"
linkTitle: "5.1 Errors are values"
weight: 1
description: "The error interface, creating errors with errors.New and fmt.Errorf, the if err != nil pattern, custom error types, and how to write good error messages."
goVersion: "1.27"
---

## What you will learn

- What the `error` type actually is
- How to create errors with `errors.New` and `fmt.Errorf`
- The `if err != nil` pattern, and why Go chose it over exceptions
- How to define your own error types with structured data
- How to write error messages that are useful in logs

## The idea

In Go, an error is just a value of the built-in interface type `error`:

```go
type error interface {
	Error() string
}
```

Anything with an `Error() string` method is an error. Functions that can fail return an `error` as their **last** result, and `nil` means success.

This is a deliberate choice. Exceptions make the failure path invisible: any line might throw, and the code that handles it can be far away. Go makes failure **explicit at every call site**. It costs some typing, and in exchange you can read a function and see exactly where it can fail and what happens when it does.

## Creating errors

{{< example file="examples/m05/values/values_test.go" region="create" run="go test -run Example_creating -v ./examples/m05/values" >}}

- `errors.New("text")` creates a fixed error.
- `fmt.Errorf(format, args...)` creates an error with a formatted message. With the `%w` verb it also **wraps** another error, which lesson 5.2 covers.

## Handling errors

{{< example file="examples/m05/values/values_test.go" region="handle" run="go test -run Example_handling -v ./examples/m05/values" >}}

The shape is always the same:

```go
v, err := doSomething()
if err != nil {
	// handle it: return it (usually with context), retry, fall back, or report it
}
// use v
```

Rules of thumb:

- **Check the error before using the other results.** When `err != nil`, the other results are usually zero values or garbage.
- **Handle it or return it, once.** Do not log an error and then also return it; the caller will log it again and you get duplicates (lesson 5.6).
- **Add context when you return it.** `parsePort` says *which* input failed. The low-level `strconv` error alone would not tell you that.
- **Keep the happy path unindented.** Return early on errors, so the main logic reads top to bottom on the left margin.

## Custom error types

When callers need more than a message, define a type:

{{< example file="examples/m05/values/values_test.go" region="custom" run="go test -run Example_customType -v ./examples/m05/values" >}}

`errors.As` checks whether an error (or anything it wraps) is a `*ValidationError` and, if so, gives you the typed value. This lets an HTTP handler, for example, return the offending field name to the client. Lesson 5.3 discusses when a type is better than a plain error value.

Use a **pointer receiver** for `Error()` on struct error types, and return them as `error` (never as `*ValidationError`), to avoid the nil interface trap from lesson 4.4.

## Writing good error messages

Go error messages are often **chained**: `load profile: find user 42: not found`. A few conventions keep the chain readable:

- **Start lowercase, no trailing punctuation.** Messages get embedded in other messages.
- **Describe what was being done**, not that "an error occurred": `open config: ...` rather than `error opening config: ...`.
- **Include the identifying values**: the file name, the ID, the input. Quote strings with `%q` so empty values and spaces are visible.
- **Do not include secrets**: passwords, tokens, full credit card numbers.

## Common mistakes

### Ignoring errors with `_`

`data, _ := os.ReadFile(path)` turns a missing file into silently empty data. If ignoring an error is truly correct, leave a comment saying why.

### Comparing error strings

`if err.Error() == "not found"` breaks when someone rewords the message or wraps it. Use `errors.Is` with a sentinel error or `errors.As` with a type (lesson 5.2).

### Capitalized messages with punctuation

`errors.New("Could not connect.")` produces `fetch: Could not connect.: timeout` once wrapped. Write `could not connect`.

### Returning a typed nil

Covered in lesson 4.4: declare results as `error` and return a literal `nil`.

## Key takeaways

- `error` is an interface with one method, `Error() string`; `nil` means success.
- Errors are returned as the last result and checked with `if err != nil`.
- Return errors with context, and handle each error once.
- Define error types when callers need structured information, and inspect them with `errors.As`.
- Messages are lowercase, without trailing punctuation, and say what was being attempted.

## Further reading

- [Error handling and Go](https://go.dev/blog/error-handling-and-go) on the Go blog
- [Errors are values](https://go.dev/blog/errors-are-values) by Rob Pike
- [Go Code Review Comments: Error strings](https://go.dev/wiki/CodeReviewComments#error-strings)
