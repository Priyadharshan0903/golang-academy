---
title: "2.1 Variables, zero values and short declarations"
linkTitle: "2.1 Variables"
weight: 1
description: "The ways to declare variables in Go, what zero values are, and the rules of the := short declaration."
goVersion: "1.27"
---

## What you will learn

- The different ways to declare a variable, and when to use each
- What **zero values** are and why they matter so much in Go
- The exact rules of `:=`
- Go's naming conventions for variables

## The idea

Go is statically typed: every variable has a type fixed at compile time. But you rarely write the type yourself, because the compiler infers it from the value.

The most important idea in this lesson is the **zero value**. In Go, there is no such thing as an uninitialized variable. Every variable you declare without a value is set to the zero value of its type. No garbage memory, no `undefined`, no compiler warning about "possibly uninitialized". This one rule shapes a lot of idiomatic Go design.

## Declaring variables

{{< example file="examples/m02/variables/variables_test.go" region="declare" run="go test -run Example_declare -v ./examples/m02/variables" >}}

Four forms, each with a typical use:

| Form | Where | Use it when |
|---|---|---|
| `var x T = v` | anywhere | You want a type different from what the value would infer, e.g. `var ratio float64 = 1` |
| `var x = v` | anywhere | Package-level variables with an initial value |
| `var x T` | anywhere | You want the zero value and will fill it in later |
| `x := v` | inside functions only | The everyday choice inside functions |

Several variables can be declared together, which is common at package level:

```go
var (
	defaultPort = 8080
	verbose     bool
)
```

## Zero values

{{< example file="examples/m02/variables/variables_test.go" region="zero" run="go test -run Example_zeroValues -v ./examples/m02/variables" >}}

| Type | Zero value |
|---|---|
| numbers (`int`, `float64`, ...) | `0` |
| `bool` | `false` |
| `string` | `""` (empty string) |
| pointers, slices, maps, channels, functions, interfaces | `nil` |
| structs | every field set to its own zero value |
| arrays | every element set to its zero value |

### Why zero values matter

Good Go types are designed so that **the zero value is useful**. You will see this everywhere in the standard library:

- A zero `strings.Builder` or `bytes.Buffer` is an empty buffer, ready to use.
- A zero `sync.Mutex` is an unlocked mutex. No constructor needed.
- A `nil` slice behaves like an empty slice: `len` is 0, you can `range` over it and `append` to it.

When you design your own types (Module 3 and 4), aim for the same property. It removes a whole class of "forgot to initialize" bugs.

{{< callout type="warning" >}}
One important exception: a `nil` **map** can be read, but writing to it panics. You must create maps with `make` or a literal before adding keys. Lesson 3.5 covers this.
{{< /callout >}}

## The rules of `:=`

`:=` declares **and** assigns, inferring the type. Its rules are short but strict:

1. **Only inside functions.** At package level you must use `var`.
2. **At least one new variable on the left.** Variables that already exist **in the same scope** are assigned, not redeclared.
3. **It always declares in the current scope.** Inside an `if` or `for` block that is a new, inner scope, which is the source of the shadowing bug in lesson 2.9.

{{< example file="examples/m02/variables/variables_test.go" region="multi" run="go test -run Example_multipleAssignment -v ./examples/m02/variables" >}}

Rule 2 is what makes this common pattern work:

```go
user, err := loadUser(id)
orders, err := loadOrders(user) // orders is new, err is reused
```

## Naming

- Use **mixedCaps**: `maxRetries`, `httpClient`. Never underscores.
- Acronyms keep one case: `userID`, `parseURL`, `HTTPServer` (not `userId` or `HttpServer`).
- **Short names for short scopes.** `i` for a loop index, `r` for a reader, `err` for an error. Longer, descriptive names for package-level variables and anything used far from where it is declared.
- Remember that an uppercase first letter exports a package-level name (lesson 1.5).

## Common mistakes

### Using `:=` at package level

```go
port := 8080 // syntax error: non-declaration statement outside function body
```

Use `var port = 8080`.

### Declaring a variable and never using it

Unused local variables are compile errors (lesson 1.3). This often happens while debugging. Delete the variable instead of silencing it with `_ = x`.

### Expecting `:=` to always create new variables

`a, err := f()` followed by `b, err := g()` reuses `err`. That is fine. But the same line inside a nested block creates a **new** `err`, which is a real bug. See lesson 2.9.

### Initializing to the zero value explicitly

`var count int = 0` or `name := ""` work, but they are noise. Prefer `var count int` and `var name string`, which say "start empty" clearly.

## Key takeaways

- Use `:=` inside functions and `var` at package level or when you want a zero value.
- Every variable starts at its type's zero value. There are no uninitialized variables.
- Design types so their zero value is useful, as the standard library does.
- `:=` needs at least one new variable on the left and always declares in the current scope.
- Use mixedCaps, consistent acronyms, and name length proportional to scope.

## Further reading

- [Variable declarations](https://go.dev/ref/spec#Variable_declarations) and [Short variable declarations](https://go.dev/ref/spec#Short_variable_declarations) in the spec
- [The zero value](https://go.dev/ref/spec#The_zero_value) in the spec
- [Effective Go: Names](https://go.dev/doc/effective_go#names)
