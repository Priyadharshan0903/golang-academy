---
title: "9.2 Test helpers, t.Cleanup, t.Context and golden files"
linkTitle: "9.2 Helpers and golden files"
weight: 2
description: "Write helpers with t.Helper, manage resources with t.Cleanup and t.TempDir, set environment safely with t.Setenv, use t.Context, keep fixtures in testdata, and compare large outputs with golden files."
goVersion: "1.27"
---

## What you will learn

- Writing helpers that report failures at the caller's line, with `t.Helper()`
- `t.TempDir`, `t.Setenv` and `t.Chdir` for safe, self-cleaning setup
- `t.Cleanup`, and how it differs from `defer`
- `t.Context` (Go 1.24) for tests of context-aware code
- The `testdata` directory and **golden files** with an `-update` flag

## Helpers

{{< example file="examples/m09/helpers/config_test.go" region="helper" run="go test -v -run TestLoadEnv ./examples/m09/helpers" >}}

- `t.Helper()` marks a function as a helper. When a helper calls `t.Fatalf`, the failure is reported at the line in the **test** that called the helper, which is where you want to look.
- Helpers take `t` (or `testing.TB`, which works for tests and benchmarks) as their first argument and call `t.Fatal` themselves, so the test reads as a clean sequence of steps.
- `t.TempDir()` returns a fresh directory that is **deleted automatically** when the test ends. No cleanup code, no leftovers, and parallel tests never collide.

The `testing` package has more of these self-cleaning helpers:

| Helper | Effect, undone automatically at the end of the test |
|---|---|
| `t.TempDir()` | create a temporary directory |
| `t.Setenv(k, v)` | set an environment variable |
| `t.Chdir(dir)` (Go 1.24) | change the working directory |
| `t.Context()` (Go 1.24) | a context cancelled when the test ends |
| `t.ArtifactDir()` (Go 1.26) | a directory for files you want to keep after the test, written under `-outputdir` |

## `t.Cleanup`

{{< example file="examples/m09/helpers/config_test.go" region="cleanup" run="go test -v -run TestCleanupOrder ./examples/m09/helpers" >}}

`t.Cleanup(f)` registers `f` to run when the test (and all its subtests) finish, in last-in, first-out order. Unlike `defer`, a **helper** can register cleanup on behalf of the test:

```go
func startServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(newHandler())
	t.Cleanup(srv.Close) // the test does not need to remember to close it
	return srv
}
```

## `t.Context`

{{< example file="examples/m09/helpers/config_test.go" region="context" run="go test -v -run TestContext ./examples/m09/helpers" >}}

Pass `t.Context()` to the code under test instead of `context.Background()`. It is cancelled just before cleanup runs, so goroutines started by the test get a signal to stop, and cleanup code can then wait for them.

## `testdata` and golden files

The `go` tool ignores any directory named **`testdata`**, so it is the conventional home for fixture files. Tests run with the package directory as the working directory, so `filepath.Join("testdata", "input.json")` works.

When output is large (rendered templates, generated code, CLI output, JSON reports), writing the expected value inline is unwieldy. A **golden file** stores the expected output in `testdata`, and an `-update` flag rewrites it:

{{< example file="examples/m09/helpers/config_test.go" region="golden" run="go test -run TestRenderGolden ./examples/m09/helpers" >}}

```bash
go test ./examples/m09/helpers                 # compare with the golden file
go test ./examples/m09/helpers -update         # accept the current output
git diff examples/m09/helpers/testdata         # review the change before committing
```

The `git diff` step is the point of golden files: changes to output show up in code review as readable diffs.

## Common mistakes

### Helpers without `t.Helper()`

Every failure points to a line inside the helper, and you have to work out which caller it was.

### `os.Setenv` in tests

It changes the environment for every test in the process and is never restored. Use `t.Setenv`.

### Writing fixtures to the package directory

Tests that create files next to the source pollute the repository and break when run in parallel. Use `t.TempDir()`.

### Running `-update` without reviewing

Accepting golden files blindly turns tests into recorders of whatever the code currently does, bugs included.

## Key takeaways

- Call `t.Helper()` in helpers; let helpers call `t.Fatal`.
- Use `t.TempDir`, `t.Setenv`, `t.Chdir` and `t.Context`: they clean up after themselves.
- `t.Cleanup` runs registered functions after the test, and helpers can use it.
- Keep fixtures in `testdata`; use golden files with an `-update` flag for large outputs, and review the diffs.

## Further reading

- [Package testing: T](https://pkg.go.dev/testing#T)
- [Go 1.24 release notes: testing](https://go.dev/doc/go1.24#testing)
