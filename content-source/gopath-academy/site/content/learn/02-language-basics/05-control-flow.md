---
title: "2.5 Control flow: if, for, switch and labels"
linkTitle: "2.5 Control flow"
weight: 5
description: "Go's small set of control structures: if with init statements, the one loop keyword, switch without fallthrough, and labeled break and continue."
goVersion: "1.27"
---

## What you will learn

- How `if` works, including the init statement idiom
- The forms of Go's only loop, `for`, including range over integers
- How `switch` differs from C and Java
- How to break out of nested loops with labels
- The `break` inside `switch` trap

## The idea

Go has very few control structures: `if`, `for`, `switch` (plus `select` for channels in Module 6, and `goto`, which you will almost never use). There is no `while`, no `do while`, no ternary `?:` operator, and no `foreach` keyword. Each of the few that exist is flexible enough to cover those cases.

Two syntax rules apply to all of them:

- **No parentheses** around conditions: `if x > 0 {`, not `if (x > 0) {`.
- **Braces are mandatory**, even for one-line bodies. There is no `if x > 0 return`.

## `if`, with an init statement

`if` can start with a short statement, separated by a semicolon. Variables declared there are scoped to the `if` and all its `else` branches:

{{< example file="examples/m02/controlflow/controlflow_test.go" region="if" run="go test -run Example_ifWithInit -v ./examples/m02/controlflow" >}}

You will see this pattern constantly with errors:

```go
if err := save(user); err != nil {
	return err
}
```

`err` exists only where it is needed, which keeps the surrounding function free of stale variables.

### No ternary operator

Go deliberately has no `cond ? a : b`. Write it out:

```go
label := "off"
if enabled {
	label = "on"
}
```

It is two extra lines, and nested ternaries are impossible.

## `for`: the only loop

{{< example file="examples/m02/controlflow/controlflow_test.go" region="for" run="go test -run Example_forLoops -v ./examples/m02/controlflow" >}}

| Form | Equivalent in other languages |
|---|---|
| `for i := 0; i < n; i++ { }` | C-style `for` |
| `for cond { }` | `while (cond)` |
| `for { }` | `while (true)` |
| `for i := range n { }` | `for i in range(n)` (Go 1.22+) |
| `for i, v := range slice { }` | `foreach` / `enumerate` |
| `for k, v := range m { }` | iterate a map (lesson 3.5) |
| `for r := range ch { }` | read a channel until closed (Module 6) |

With `range`, drop what you do not need: `for i := range s` for indexes only, `for _, v := range s` for values only.

Two details about `range`:

- The value is a **copy** of the element. Assigning to `v` does not change the slice. Use the index (`s[i] = ...`) to modify elements.
- **Map iteration order is random** and changes between runs on purpose. Never rely on it.

Since Go 1.22, each loop iteration gets **fresh** loop variables. This fixed a famous bug with closures and goroutines capturing the loop variable, which you will see in lesson 2.6. Old blog posts that warn about it describe behavior before Go 1.22.

## `switch`

{{< example file="examples/m02/controlflow/controlflow_test.go" region="switch" run="go test -run Example_switch -v ./examples/m02/controlflow" >}}

Differences from C, Java and JavaScript:

- **No automatic fallthrough.** Each case ends on its own, so no `break` is needed.
- **A case can list several values**: `case 200, 201, 204:`.
- **Cases do not have to be constants.** They can be any expression, evaluated top to bottom.
- **A switch with no expression** is a clean replacement for long `if / else if` chains, as in `grade` above.
- Like `if`, a `switch` can have an init statement: `switch n := compute(); n { ... }`.

If you really want to continue into the next case, say so with `fallthrough`. It is rare:

{{< example file="examples/m02/controlflow/controlflow_test.go" region="fallthrough" run="go test -run Example_fallthrough -v ./examples/m02/controlflow" >}}

`fallthrough` runs the next case's body **without** checking its condition. There is also a **type switch** for checking the dynamic type of an interface value, covered in lesson 4.3.

## Labels: breaking out of nested loops

`break` and `continue` apply to the innermost loop. To affect an outer loop, put a **label** on it:

{{< example file="examples/m02/controlflow/controlflow_test.go" region="labels" run="go test -run Example_labels -v ./examples/m02/controlflow" >}}

`continue outer` works the same way, skipping to the next iteration of the labeled loop.

## The `break` inside `switch` trap

Inside a `switch` (or `select`), a plain `break` exits the **switch**, not the surrounding loop. This surprises almost everyone once:

{{< example file="examples/m02/controlflow/controlflow_test.go" region="breakswitch" run="go test -run Example_breakInsideSwitch -v ./examples/m02/controlflow" >}}

When you need to leave a loop from inside a `switch` or `select`, use a labeled `break`. You will need this often with `select` in concurrent code.

## Common mistakes

### Modifying the range value instead of the element

```go
for _, p := range players {
	p.Score = 0 // changes a copy; players is untouched
}
```

Use `for i := range players { players[i].Score = 0 }`.

### Relying on map iteration order

Tests that pass locally and fail in CI often do so because of map order. Sort the keys first (`slices.Sorted(maps.Keys(m))`, lesson 3.4) when order matters.

### Plain `break` inside `switch` or `select` in a loop

It only leaves the switch. Use a label.

### Adding `break` at the end of every case

Harmless but noisy, and a sign of C or Java habits. Go cases end on their own.

## Key takeaways

- No parentheses, mandatory braces, no ternary operator.
- `if` and `switch` can take an init statement that scopes variables tightly.
- `for` is the only loop and covers C-style, while, infinite and range loops, including `range n` since Go 1.22.
- `switch` cases do not fall through, can list several values, and a bare `switch` replaces `if / else if` chains.
- Use labels to `break` or `continue` an outer loop, especially from inside `switch` or `select`.

## Further reading

- [Effective Go: Control structures](https://go.dev/doc/effective_go#control-structures)
- [For statements](https://go.dev/ref/spec#For_statements) and [Switch statements](https://go.dev/ref/spec#Switch_statements) in the spec
- [Fixing for loops in Go 1.22](https://go.dev/blog/loopvar-preview) on the Go blog
- [Range over integers](https://go.dev/doc/go1.22#language) in the Go 1.22 release notes
