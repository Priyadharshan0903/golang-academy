---
title: "9.1 Table-driven tests and subtests"
linkTitle: "9.1 Table-driven tests"
weight: 1
description: "Write tests as tables of cases with t.Run subtests, write failure messages that explain themselves, run tests in parallel, select tests with -run, and check that your tests actually catch bugs."
goVersion: "1.27"
---

## What you will learn

- The anatomy of a Go test, and `t.Error` versus `t.Fatal`
- The table-driven pattern with `t.Run` subtests
- How to write failure messages that need no debugger
- Running subtests in parallel
- Selecting tests with `-run` and subtest names
- How to tell whether your tests are good enough: **mutation testing**

## The basics

A test is a function `func TestXxx(t *testing.T)` in a file ending with `_test.go`. `go test` finds, compiles and runs them. A test passes unless it calls one of the failure methods:

| Method | Effect |
|---|---|
| `t.Error`, `t.Errorf` | Mark the test failed and **keep going** |
| `t.Fatal`, `t.Fatalf` | Mark it failed and **stop this test** now |
| `t.Log`, `t.Logf` | Print only if the test fails, or with `-v` |
| `t.Skip`, `t.Skipf` | Skip the test with a reason |

Use `Error` when later checks still make sense, so one run reports every problem. Use `Fatal` when continuing is pointless (setup failed, a value you need is nil). There are no assertion functions in the standard library; plain `if` statements with good messages are the Go style.

## Table-driven tests

Here is a function to test:

{{< example file="examples/m09/tables/slug.go" region="code" >}}

And its test, in the form you will see in almost every Go codebase:

{{< example file="examples/m09/tables/slug_test.go" region="table" run="go test -v -run 'TestSlugify$' ./examples/m09/tables" >}}

1. A slice of anonymous structs, one per case (lesson 3.6).
2. A loop over the cases.
3. `t.Run(name, func)` runs each case as a **subtest** with its own name, result and output.

Adding a case is one line. Reading the table tells you what the function is supposed to do. And because each case is a subtest, the output names the failing case:

```text
--- FAIL: TestSlugify (0.00s)
    --- FAIL: TestSlugify/unicode_letters (0.00s)
        slug_test.go:30: Slugify("Café Olé") = "caf-ol", want "café-olé"
```

## Failure messages

A good failure message lets someone fix the bug without rerunning anything. The Go convention:

```text
FunctionName(inputs) = got, want expected
```

- Include the **inputs**, so the failure can be reproduced.
- Print **got before want**.
- Use `%q` for strings so empty strings and whitespace are visible.
- Say what was being checked when it is not obvious: `status = 500, want 200`.

## Parallel subtests

{{< example file="examples/m09/tables/slug_test.go" region="parallel" run="go test -v -run TestSlugifyParallel ./examples/m09/tables" >}}

`t.Parallel()` lets a subtest run at the same time as other parallel subtests. Use it for tests that are slow because they wait (network calls to a test server, sleeps). Since Go 1.22, the loop variable `tt` is per iteration, so capturing it in the closure is safe; older code had a `tt := tt` line for this.

Parallel tests must not share mutable state, and `t.Setenv` and `t.Chdir` cannot be used in them because they change process-wide state.

## Selecting tests

```bash
go test ./...                                     # everything
go test -run TestSlugify ./examples/m09/tables    # regexp on test names
go test -run 'TestSlugify/unicode' ./...          # one subtest: spaces in names become _
go test -run TestSlugify -count=1 ./...           # bypass the test cache
go test -v -failfast ./...                        # stop at the first failure
```

`-count=1` matters when a test depends on something outside the code (files, environment, time), because `go test` caches passing results.

## Are your tests any good?

Coverage (lesson 9.8) tells you which lines ran, not whether your tests would **notice a bug** in them. A direct way to check: plant bugs in the code and see if the tests fail. This is called **mutation testing**. Tools can do it automatically; the exercise below does it by hand.

## Exercise: test your tests

<div class="ga-exercise">
  <div class="ga-exercise-title">Exercise</div>
  <div class="ga-exercise-body">

This exercise is reversed: the code is given and correct, and **you write the tests**.

In [`exercises/m09/wordwrap`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/exercises/m09/wordwrap), `wrap.go` contains a correct `Wrap(text, width)` function. Next to it are four files, `bug_overflow.go`, `bug_longword.go`, `bug_empty.go` and `bug_lastline.go`, each a subtly broken version selected with a build tag. Read the doc comment on `Wrap`, then extend `wrap_test.go` into a table-driven test that:

1. **passes** against the correct code, and
2. **fails** against every buggy version.

```bash
cd exercises
go test ./m09/wordwrap                       # must PASS
for b in bug_overflow bug_longword bug_empty bug_lastline; do
    go test -tags $b ./m09/wordwrap > /dev/null && echo "$b NOT caught"
done                                          # must print nothing
```

The starting test already catches one bug. Try to catch the rest **from the specification**, without reading the bug files first. The reference tests are in `solution_test.go` (build tag `solution`), and CI checks that they catch every bug with `scripts/check-mutants.sh`.

  </div>
</div>

## Common mistakes

### Messages without inputs

`t.Error("wrong result")` forces whoever sees the failure to rerun it under a debugger.

### `t.Fatal` inside a goroutine

`Fatal` must be called from the goroutine running the test. In other goroutines, use `t.Error` and return, or send the error back over a channel.

### One giant test function

A 200-line test that checks everything in sequence stops at the first `Fatal` and hides other failures. Split it into subtests.

### Testing implementation details

Tests that check private fields or the exact sequence of internal calls break on every refactor. Test behavior through the public API, as an external test package (`package foo_test`) forces you to (lesson 1.5).

## Key takeaways

- Tests are `TestXxx(t *testing.T)` functions in `_test.go` files; use `Error` to continue and `Fatal` to stop.
- Table-driven tests with `t.Run` are the standard shape: one line per case, one subtest per case.
- Failure messages show inputs, got, then want.
- `t.Parallel()` speeds up waiting tests; `-run` selects tests and subtests.
- Good tests catch bugs: try planting some and see.

## Further reading

- [Package testing](https://pkg.go.dev/testing)
- [Go wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests)
- [Using Subtests and Sub-benchmarks](https://go.dev/blog/subtests) on the Go blog
- [Go Code Review Comments: Useful Test Failures](https://go.dev/wiki/CodeReviewComments#useful-test-failures)
