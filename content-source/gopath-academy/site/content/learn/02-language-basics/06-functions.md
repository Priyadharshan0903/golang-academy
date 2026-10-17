---
title: "2.6 Functions: multiple returns, variadics and closures"
linkTitle: "2.6 Functions"
weight: 6
description: "Declaring functions, returning several values, named results, variadic parameters, functions as values, and closures."
goVersion: "1.27"
---

## What you will learn

- How Go functions are declared, and why there is no overloading or default arguments
- Multiple return values and the `(value, error)` convention
- Named results, and when bare `return` is a good or bad idea
- Variadic functions and spreading a slice into them
- Functions as values, and closures that capture variables

## Declaring functions

```go
func add(a int, b int) int { return a + b }
func add2(a, b int) int    { return a + b } // same type: group the names
```

The type comes **after** the name, for parameters and results alike. Go reads left to right: "add takes a and b, both int, and returns an int".

What Go deliberately leaves out:

- **No overloading.** Two functions cannot share a name with different parameters. Use different names: `strings.Split` and `strings.SplitN`.
- **No default parameter values.** Pass the value explicitly, or for many optional settings use an options struct (lesson 4.8 and Module 12).
- **No keyword arguments.** Struct literals with field names (`Config{Port: 8080}`) fill that role.

These omissions keep every call site unambiguous: you can read a call and know exactly which function runs with which values.

## Multiple return values

A function can return several values. The most common use is a result plus an error:

{{< example file="examples/m02/functions/functions_test.go" region="multi" run="go test -run Example_multipleReturns -v ./examples/m02/functions" >}}

The convention is strong and you should follow it:

- The `error` is always the **last** result.
- When `err != nil`, the other results are usually zero values and should not be used.
- Check the error immediately after the call.

Other common pairs are `(value, ok bool)` for "found or not" (you will see it with maps and type assertions) and `(value, error)` for anything that can fail. Error handling gets a whole module (Module 5).

## Named results

Results can have names. They are declared at the top of the function, start at their zero values, and a bare `return` returns their current values:

{{< example file="examples/m02/functions/functions_test.go" region="named" run="go test -run Example_namedResults -v ./examples/m02/functions" >}}

Named results are useful for two things:

1. **Documentation.** `func split(full string) (first, last string)` tells the reader what each string means.
2. **Changing results from a `defer`**, covered in lesson 2.7.

Bare `return` in a **short** function is fine. In a long function it hides what is being returned. A good rule: name results when it helps documentation, but write `return first, last` explicitly unless the function fits on one screen.

## Variadic functions

A final parameter of type `...T` accepts any number of `T` arguments. Inside the function it is a slice `[]T`:

{{< example file="examples/m02/functions/functions_test.go" region="variadic" run="go test -run Example_variadic -v ./examples/m02/functions" >}}

To pass an existing slice, spread it with `values...`. You have already used a variadic function many times: `fmt.Println(a ...any)`.

## Functions are values

Functions are first-class values. You can assign them to variables, pass them as arguments, return them and store them in structs or maps. A function's type is written like its signature: `func(int) int`.

{{< example file="examples/m02/functions/functions_test.go" region="values" run="go test -run Example_functionValues -v ./examples/m02/functions" >}}

The second call uses a **function literal** (an anonymous function) written inline. This style is common for small callbacks, sorting comparisons (`slices.SortFunc`), HTTP handlers and test helpers.

The zero value of a function type is `nil`. Calling a `nil` function panics.

## Closures

A function literal can use variables from the surrounding function. It **captures the variable itself**, not a copy of its current value, and keeps it alive as long as the closure exists:

{{< example file="examples/m02/functions/functions_test.go" region="closures" run="go test -run Example_closures -v ./examples/m02/functions" >}}

Each call to `counter()` creates a new `count` variable, so `next` and `other` are independent. Closures are how Go does a lot of things other languages use classes or decorators for: generators, middleware, retry wrappers and lazy initialization.

### Closures in loops (fixed in Go 1.22)

Before Go 1.22, a `for` loop had **one** loop variable shared by all iterations, so closures created in the loop all saw its final value. A classic bug. Since Go 1.22, each iteration has its own variable:

{{< example file="examples/m02/functions/functions_test.go" region="loopvar" run="go test -run Example_loopVariables -v ./examples/m02/functions" >}}

This applies to modules whose `go.mod` says `go 1.22` or later. If you maintain older code with an earlier `go` line, the old behavior still applies there until you raise it.

## Common mistakes

### Ignoring a returned error

`result, _ := strconv.Atoi(input)` throws away the only signal that `result` is garbage. Handle the error, or at least make ignoring it a deliberate, commented decision.

### Long functions with bare returns

Readers have to scroll up to find out what is being returned and whether it was set on this path. Return values explicitly.

### Simulating overloading with `...any`

`func Do(args ...any)` that inspects its arguments at run time loses all compile-time checking. Write separate, clearly named functions.

### Calling a `nil` function value

An unset callback field panics when called. Check `if f != nil` or give it a no-op default.

## Exercise

{{< exercise dir="m02/stats" >}}
Implement `Sum` (variadic), `MinMax` (multiple results with an `ok` flag) and `RunningAverage`, which returns a closure that remembers every value it has seen. The tests check that two running averages do not share state.
{{< /exercise >}}

## Key takeaways

- Types come after names; there is no overloading, no default arguments and no keyword arguments.
- Return `(value, error)` with the error last, and check it right away.
- Name results for documentation; use bare `return` only in short functions.
- `...T` collects arguments into a `[]T`; spread a slice with `s...`.
- Functions are values, and closures capture variables. Since Go 1.22 each loop iteration has its own loop variable.

## Further reading

- [Function declarations](https://go.dev/ref/spec#Function_declarations) and [Function literals](https://go.dev/ref/spec#Function_literals) in the spec
- [Effective Go: Functions](https://go.dev/doc/effective_go#functions)
- [Fixing for loops in Go 1.22](https://go.dev/blog/loopvar-preview)
- [Codewalk: First-Class Functions in Go](https://go.dev/doc/codewalk/functions/)
