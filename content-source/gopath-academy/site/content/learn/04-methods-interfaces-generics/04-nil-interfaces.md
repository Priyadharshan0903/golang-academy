---
title: "4.4 Nil interfaces vs interfaces holding nil"
linkTitle: "4.4 Nil interfaces"
weight: 4
description: "Why an interface that holds a nil pointer is not equal to nil, how this breaks error checks, and the simple rule that avoids it."
goVersion: "1.27"
---

## What you will learn

- Why `err != nil` can be true even though "nothing went wrong"
- The (type, value) pair inside every interface value
- The rule that prevents the bug
- How to check for a typed nil when you really need to

## The bug

This is one of the most famous Go gotchas, and it almost always involves `error`:

{{< example file="examples/m04/nilinterface/nilinterface_test.go" region="trap" run="go test -run Example_nilInterfaceTrap -v ./examples/m04/nilinterface" >}}

`validateBroken(true)` returns a value that **is not equal to `nil`**, even though the pointer inside it is nil. Every caller that writes `if err != nil` thinks validation failed.

## Why it happens

Recall from lesson 4.2 that an interface value is a pair: **(dynamic type, value)**. An interface is `nil` only when **both** are unset.

{{< example file="examples/m04/nilinterface/nilinterface_test.go" region="pair" run="go test -run Example_typeValuePair -v ./examples/m04/nilinterface" >}}

| Interface value | Dynamic type | Value | `== nil` |
|---|---|---|---|
| `var e1 error` | none | none | **true** |
| `var e2 error = (*MyError)(nil)` | `*MyError` | nil pointer | **false** |
| `var e3 error = &MyError{"x"}` | `*MyError` | pointer | false |

In `validateBroken`, `err` is a variable of type `*MyError`. Returning it as an `error` converts it to an interface value whose type is `*MyError`. The type is set, so the interface is not nil, regardless of the pointer.

## The rule

**Never return a concrete pointer type through an interface when you mean "nothing".** Return a literal `nil`:

```go
func validate(ok bool) error {
	if !ok {
		return &MyError{Msg: "invalid"}
	}
	return nil
}
```

Related habits that keep you safe:

- Declare functions to return `error`, not `*MyError`. Functions returning concrete error types invite this bug at every call site that stores the result in an `error` variable.
- Do not keep a typed error variable around and return it at the end. Return `nil` explicitly on the success path.
- The same applies to any interface, not only `error`: returning a nil `*Logger` as a `Logger` interface produces a non-nil interface whose method calls panic.

## Detecting a typed nil

You rarely need to, but sometimes you must handle values from code you do not control. `errors.As` finds the typed nil (the second line of the pair example), after which you can compare the pointer to `nil`. For arbitrary interfaces, the `reflect` package can check `reflect.ValueOf(v).IsNil()` for pointer kinds. If you find yourself doing this often, fix the code that produces typed nils instead.

## Common mistakes

### Returning `*MyError` from a function

The function signature itself is the bug waiting to happen. Return `error`.

### A shared `var err *MyError` in a long function

Assigning only in failure branches and returning `err` at the end produces a typed nil on success.

### Assuming `fmt.Println(err)` reveals the problem

It prints `<nil>` for a typed nil, exactly like a real nil error, which makes the bug confusing to debug. Print with `%T` to see the dynamic type.

## Key takeaways

- An interface value is `nil` only if both its dynamic type and its value are unset.
- Storing a nil pointer in an interface gives a non-nil interface.
- Return a literal `nil` for "no error", and declare results as `error`, not concrete error types.
- `%T` in `fmt` shows the dynamic type when debugging.

## Further reading

- [Go FAQ: Why is my nil error value not equal to nil?](https://go.dev/doc/faq#nil_error)
- [The Laws of Reflection](https://go.dev/blog/laws-of-reflection)
