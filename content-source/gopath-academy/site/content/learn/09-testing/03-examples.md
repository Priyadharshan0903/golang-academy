---
title: "9.3 Testable examples"
linkTitle: "9.3 Testable examples"
weight: 3
description: "Write Example functions that are both documentation on pkg.go.dev and tests run by go test, with naming rules, Output and Unordered output."
goVersion: "1.27"
---

## What you will learn

- How `Example` functions become documentation and tests at the same time
- The naming rules that attach examples to packages, types and methods
- `// Output:` and `// Unordered output:`
- When an example is better than a unit test, and when it is not

## The idea

You have been reading testable examples since lesson 1.3: every example in this course is one. An `Example` function in a `_test.go` file is:

1. **Compiled and run by `go test`**, with its standard output compared to the `// Output:` comment.
2. **Shown in the documentation**, by `go doc` and on [pkg.go.dev](https://pkg.go.dev), next to whatever it documents.

So examples are documentation that cannot go out of date.

## Naming

{{< example file="examples/m09/examplesdoc/example_test.go" region="names" run="go test -v ./examples/m09/examplesdoc" >}}

| Function name | Documents |
|---|---|
| `Example` | the package |
| `ExampleF` | the function `F` |
| `ExampleT` | the type `T` |
| `ExampleT_M` | the method `M` on type `T` |
| any of the above plus `_suffix` | an additional example; the suffix starts with a lowercase letter |

Examples are usually written in the external test package (`package examplesdoc_test`), so they show the API exactly as users call it, with the package name in front.

## Output checks

- `// Output:` followed by the expected lines. Leading and trailing whitespace of the whole output is trimmed before comparing (which is why trailing spaces inside lines bit us in earlier modules).
- `// Unordered output:` compares lines in any order, for output from maps or goroutines:

{{< example file="examples/m09/examplesdoc/example_test.go" region="unordered" run="go test -v -run Example_unordered ./examples/m09/examplesdoc" >}}

- An example **without** an output comment is compiled but not run. That is useful for examples that need a network or print something machine-dependent, though they then only prove the code compiles.

## Examples or unit tests?

Use examples for **showing how to use an API**: the common path, written the way you want users to write it. Use unit tests for **edge cases, error paths and exhaustive tables**. An example with twenty asserts is a bad example; a test without readable usage is a bad tutorial. Most good packages have both.

## Common mistakes

### Machine-dependent output

Printing times, map iteration, pointers or goroutine timing makes examples flaky. Sort, round or use `Unordered output`.

### Wrong names

`ExampleStack_pop` (lowercase method name) does not attach to the `Pop` method, and `go vet` warns about examples that refer to unknown identifiers.

### Examples that hide errors

`v, _ := parse(...)` teaches readers to ignore errors. Show proper handling, even in examples.

## Key takeaways

- `Example` functions are run by `go test` and shown in the documentation.
- Names attach them to the package, functions, types and methods; suffixes add more examples.
- `// Output:` checks output exactly; `// Unordered output:` ignores order.
- Use examples to teach the API and unit tests to cover the edge cases.

## Further reading

- [Testable Examples in Go](https://go.dev/blog/examples) on the Go blog
- [Package testing: Examples](https://pkg.go.dev/testing#hdr-Examples)
