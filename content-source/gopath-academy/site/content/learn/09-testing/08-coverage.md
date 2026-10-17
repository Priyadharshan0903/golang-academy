---
title: "9.8 Coverage and integration tests"
linkTitle: "9.8 Coverage and integration"
weight: 8
description: "Measure test coverage and read coverage reports, understand what coverage does not tell you, separate fast unit tests from integration tests, and use TestMain for shared setup."
goVersion: "1.27"
---

## What you will learn

- Measuring coverage with `-cover` and viewing it as HTML
- Why 100% coverage does not mean the code is tested
- Coverage for integration tests of whole binaries
- Separating slow tests with `-short`, environment variables and build tags
- Shared setup with `TestMain`

## Measuring coverage

{{< example file="examples/m09/coverage/discount.go" region="code" >}}

{{< example file="examples/m09/coverage/discount_test.go" region="tests" >}}

```bash
go test -cover ./examples/m09/coverage
```

```text
ok  	.../examples/m09/coverage	0.003s	coverage: 100.0% of statements
```

```bash
go test -coverprofile=cover.out ./...   # write a profile
go tool cover -func=cover.out           # per-function percentages
go tool cover -html=cover.out           # open an annotated view of the source in a browser
```

The HTML view colors each statement green (ran) or red (never ran), which is the fastest way to spot untested branches.

## What coverage does not tell you

The test above reaches **100%** of statements. Yet it never checks a member using the `WELCOME` coupon, where the inner `if d == 0` is false. Statement coverage records that the `if` line ran, not that both outcomes were tested. It also says nothing about whether the assertions are any good: a test that calls every function and checks nothing has full coverage.

Use coverage as a **flashlight** for finding untested code, not as a target. Lesson 9.1's mutation exercise is a much stronger check of test quality.

## Coverage for whole programs

Since Go 1.20 you can build a **binary** with coverage instrumentation and collect coverage from integration tests or manual runs:

```bash
go build -cover -o app ./cmd/app
GOCOVERDIR=/tmp/cov ./app --some-flags      # each run writes coverage data
go tool covdata percent -i=/tmp/cov
go tool covdata textfmt -i=/tmp/cov -o cover.out   # convert for `go tool cover`
```

This answers "which code did our end-to-end test suite actually exercise?".

## Separating slow tests

{{< example file="examples/m09/coverage/discount_test.go" region="short" run="go test -short -v ./examples/m09/coverage" >}}

Three common ways to keep the default `go test ./...` fast:

1. **`-short`**: slow tests call `t.Skip` when `testing.Short()` is true. Run `go test -short ./...` in the edit loop.
2. **Environment variables**: tests that need external systems skip unless their configuration is set (`PRICING_URL`, `DATABASE_URL`). CI sets them in a dedicated job.
3. **Build tags**: put integration tests in files starting with `//go:build integration` and run them with `go test -tags integration ./...`.

Pick one convention per project and document it in the README.

## `TestMain`

{{< example file="examples/m09/coverage/main_test.go" region="testmain" >}}

`TestMain` runs once per test binary (once per package) and must call `m.Run()`. Use it for expensive shared setup: starting a database container, building a binary under test, loading large fixtures. Most packages do not need it; `t.Cleanup` and helpers are enough.

## Common mistakes

### Coverage targets

"Coverage must be 80%" leads to tests that execute code without checking it. Review what is tested, not just the number.

### Integration tests in the default run

Tests that need Docker or network access make `go test ./...` fail for new contributors. Gate them.

### Forgetting `os.Exit(m.Run())`

A `TestMain` that does not exit with `m.Run()`'s code makes failing tests look like passes.

## Key takeaways

- `go test -cover`, `-coverprofile` and `go tool cover -html` show what ran.
- Full statement coverage does not mean every branch or behavior is checked.
- `go build -cover` plus `GOCOVERDIR` measures coverage of whole-program runs.
- Keep slow tests out of the default run with `-short`, environment variables or build tags.
- `TestMain` is for rare, expensive shared setup.

## Further reading

- [The cover story](https://go.dev/blog/cover) on the Go blog
- [Coverage profiling support for integration tests](https://go.dev/doc/build-cover)
- [Package testing: Main](https://pkg.go.dev/testing#hdr-Main)
