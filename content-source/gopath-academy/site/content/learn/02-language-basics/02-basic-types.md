---
title: "2.2 Basic types, numeric conversions and overflow"
linkTitle: "2.2 Basic types"
weight: 2
description: "Go's built-in types, why every conversion is explicit, how integer overflow and division behave, and how to convert between strings and numbers."
goVersion: "1.27"
---

## What you will learn

- The built-in types and which ones to use by default
- Why Go never converts between numeric types automatically
- How integer overflow, division and float comparison really behave
- How to convert between numbers and strings correctly

## The built-in types

| Kind | Types | Default choice |
|---|---|---|
| Boolean | `bool` | |
| Text | `string` | |
| Signed integers | `int`, `int8`, `int16`, `int32`, `int64` | `int` |
| Unsigned integers | `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr` | rarely needed |
| Aliases | `byte` = `uint8`, `rune` = `int32` | `byte` for raw data, `rune` for characters |
| Floating point | `float32`, `float64` | `float64` |
| Complex | `complex64`, `complex128` | rarely needed |

A few things to know:

- **`int` and `uint` are 32 or 64 bits** depending on the platform. On every 64-bit system you are likely to deploy to, `int` is 64 bits.
- **Use `int` by default** for counts, indexes and sizes. `len()` returns an `int`, and mixing types costs you conversions everywhere.
- **Use sized types** when the size is part of a contract: file formats, network protocols, database columns, or large arrays where memory matters.
- **Avoid unsigned types for quantities** that you subtract. `uint` underflow wraps to a huge number instead of going negative (shown below).
- `byte` and `rune` are aliases, not new types. `byte` and `uint8` are completely interchangeable.

## Conversions are always explicit

Go has no implicit numeric conversion. Not even from `int32` to `int64`, and not from `int` to `float64`. You write the conversion as `T(value)`:

{{< example file="examples/m02/numbers/numbers_test.go" region="conversion" run="go test -run Example_conversion -v ./examples/m02/numbers" >}}

This feels verbose at first. It exists because implicit conversions are a classic source of bugs in C and Java: silent precision loss, sign changes and surprising results in mixed expressions. In Go, every place a value changes type is visible in the code.

Even two types with the same underlying representation are different:

```go
var a int = 1
var b int64 = 2
fmt.Println(a + b) // invalid operation: a + b (mismatched types int and int64)
```

## Overflow wraps at run time

Integer arithmetic that exceeds the type's range **wraps around** silently at run time. There is no exception and no panic:

{{< example file="examples/m02/numbers/numbers_test.go" region="overflow" run="go test -run Example_overflow -v ./examples/m02/numbers" >}}

Constants are different. The compiler checks constant expressions, so this fails to build:

```go
var x int8 = 128 // cannot use 128 (untyped int constant) as int8 value in variable declaration (overflows)
```

Use `math.MaxInt64`, `math.MinInt32` and friends from the `math` package when you need the limits. If overflow is a real risk (money, counters that run for years), use `int64` and check before adding, or use `math/big` for arbitrary precision.

## Division and remainder

- Integer division **truncates toward zero**: `7 / 2` is `3`, and `-7 / 2` is `-3`.
- `%` has the sign of the dividend: `-7 % 2` is `-1`.
- Dividing an integer by zero **panics** at run time (or is a compile error with constants). Dividing a float by zero gives `+Inf`, `-Inf` or `NaN`.

## Floating point

`float64` follows IEEE 754, like every mainstream language. That means the usual surprises:

{{< example file="examples/m02/numbers/numbers_test.go" region="floats" run="go test -run Example_floats -v ./examples/m02/numbers" >}}

Never compare floats with `==` after arithmetic. And **never use floats for money**: store amounts as integer minor units (paise or cents) in an `int64`, or use a decimal library.

## Numbers and strings

The `strconv` package converts between numbers and their text form:

{{< example file="examples/m02/numbers/numbers_test.go" region="strconv" run="go test -run Example_strconv -v ./examples/m02/numbers" >}}

| Task | Function |
|---|---|
| string to int | `strconv.Atoi(s)` |
| int to string | `strconv.Itoa(n)` |
| string to int64 with base and size | `strconv.ParseInt(s, 10, 64)` |
| string to float | `strconv.ParseFloat(s, 64)` |
| string to bool | `strconv.ParseBool(s)` |
| number to string with formatting | `strconv.FormatFloat`, or `fmt.Sprintf("%.2f", f)` |

Every parse function returns an error. Always check it: user input is never as clean as you hope.

## Common mistakes

### `string(n)` to convert a number to text

`string(65)` is `"A"`, not `"65"`, because converting an integer to a string interprets it as a Unicode code point. `go vet` warns about this for non-rune integer types. Use `strconv.Itoa`.

### Unsigned subtraction

```go
var stock uint = 3
stock -= 5 // 18446744073709551614, not -2
```

Use `int` for anything that can conceptually go below zero, even if you think it never will.

### Floats for money or exact comparisons

Rounding errors accumulate. Use integers of the smallest currency unit, and compare floats with a tolerance.

### Converting a large `int64` to `int32` without checking

The conversion silently keeps the low 32 bits. Check the range first when the value comes from outside your program.

## Key takeaways

- Default to `int`, `float64`, `string`, `bool`, `byte` and `rune`.
- Every numeric conversion is explicit, even between integer sizes.
- Integer overflow wraps silently at run time; constant overflow is a compile error.
- Integer division truncates toward zero; integer division by zero panics.
- Compare floats with a tolerance, keep money in integers, and use `strconv` (not `string(n)`) for text conversion.

## Further reading

- [Numeric types](https://go.dev/ref/spec#Numeric_types) and [Conversions](https://go.dev/ref/spec#Conversions) in the spec
- [Arithmetic operators](https://go.dev/ref/spec#Arithmetic_operators), including integer overflow
- [Package strconv](https://pkg.go.dev/strconv)
- [Package math](https://pkg.go.dev/math) for limits and functions
