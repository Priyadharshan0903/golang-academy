---
title: "2.3 Constants and iota"
linkTitle: "2.3 Constants and iota"
weight: 3
description: "Typed and untyped constants, how iota builds enumerations and bit flags, and the traps of changing enum values."
goVersion: "1.27"
---

## What you will learn

- What constants can and cannot be in Go
- The difference between **typed** and **untyped** constants, and why untyped ones are so convenient
- How `iota` creates enumerations, sizes and bit flags
- How to design enums that are safe to change

## The idea

A constant is a value fixed at compile time. In Go, constants can only be **booleans, numbers, strings or runes**. There are no constant slices, maps or structs; use a package-level `var` for those.

Go constants have a feature most languages lack: they can be **untyped**, which lets one constant work naturally with many types.

## Untyped constants

`const timeout = 5` has no fixed type. It is just "the number 5". It takes on a type only when it is used, which means it fits wherever a 5 would fit:

{{< example file="examples/m02/constants/constants_test.go" region="untyped" run="go test -run Example_untyped -v ./examples/m02/constants" >}}

Two things make untyped constants powerful:

1. **No conversions needed.** `timeout * time.Second` works even though `time.Second` is a `time.Duration`. If `timeout` were an `int` variable, you would have to write `time.Duration(timeout) * time.Second`.
2. **Arbitrary precision.** Constant arithmetic happens at compile time with at least 256 bits of precision, so `1 << 100` is fine as an intermediate value. It only has to fit when it is finally assigned to a typed variable.

A **typed** constant behaves like a value of that type everywhere:

```go
const maxUsers int = 100
var limit int64 = maxUsers // error: cannot use maxUsers (constant 100 of type int) as int64 value in variable declaration
```

Prefer untyped constants unless you specifically want to restrict how the constant is used, as with enums below.

## `iota`: enumerations

Inside a `const ( ... )` block, `iota` is a counter: `0` on the first line, `1` on the second, and so on. If a line has no expression, it **repeats the previous expression** with the new value of `iota`.

{{< example file="examples/m02/constants/constants_test.go" region="iota" run="go test -run Example_iota -v ./examples/m02/constants" >}}

Defining a named type (`type Weekday int`) and giving the constants that type means a function can require a `Weekday`, and the compiler stops you from passing an arbitrary unrelated value like a `Month`. Named types are covered properly in Module 3. In Module 4 you will add a `String()` method so `Wednesday` prints as `"Wednesday"` instead of `3`; the `stringer` tool from lesson 1.4 can generate it.

## `iota` in expressions

Because the expression repeats, `iota` can build sequences that are not just 0, 1, 2:

{{< example file="examples/m02/constants/constants_test.go" region="iotaexpr" run="go test -run Example_iotaExpressions -v ./examples/m02/constants" >}}

- `_ = iota` throws away the first value, so `KB` starts at `iota == 1`.
- `1 << (10 * iota)` gives 2<sup>10</sup>, 2<sup>20</sup>, 2<sup>30</sup>.
- `1 << iota` gives one bit per constant, so flags can be combined with `|` and tested with `&`.

`iota` restarts at `0` in every new `const` block.

## Designing safe enums

### Make the zero value mean "unknown"

A variable of type `Weekday` that was never set holds `0`, which is `Sunday` above. That is a silent, plausible wrong value. For most enums, reserve `0` for "not set":

```go
const (
	StatusUnknown Status = iota // the zero value
	StatusActive
	StatusSuspended
)
```

Now a missing value is visibly `StatusUnknown` instead of quietly being the first real option.

### Never reorder or insert in the middle

`iota` values depend on **position**. If the numbers are stored in a database, sent over the network or written to a file, inserting a new constant in the middle changes the meaning of every stored value after it. Always **append** new values at the end, or assign explicit numbers when values are persisted.

## Common mistakes

### Expecting constant arrays or maps

`const days = []string{"Mon", "Tue"}` does not compile. Use `var`, and if you want callers not to modify it, keep it unexported and expose a function that returns a copy.

### Inserting a new enum value in the middle

Covered above, and it is the most damaging enum mistake because nothing fails at compile time.

### Using the first enum value as a meaningful default

If the zero value is a real option, you cannot tell "chosen" from "never set". Start with an `Unknown` or `Invalid` member.

### Assuming `iota` continues across blocks

Every `const` block starts over at `0`.

## Key takeaways

- Constants are compile-time values of boolean, numeric or string kind only.
- Untyped constants adapt to the type of their context and are computed with high precision.
- `iota` counts lines in a `const` block, and an omitted expression repeats the previous one.
- Use a named type for enums, reserve the zero value for "unknown", and only append new values.

## Further reading

- [Constants](https://go.dev/blog/constants) on the Go blog, an excellent deep dive
- [Constant declarations](https://go.dev/ref/spec#Constant_declarations) and [Iota](https://go.dev/ref/spec#Iota) in the spec
