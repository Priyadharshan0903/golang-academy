---
title: "9.7 Benchmarks with b.Loop"
linkTitle: "9.7 Benchmarks"
weight: 7
description: "Write benchmarks with testing.B and the b.Loop method from Go 1.24, report allocations, structure sub-benchmarks, read the output, and turn allocation budgets into tests."
goVersion: "1.27"
---

## What you will learn

- How to write a benchmark, and why `b.Loop` (Go 1.24) is the modern form
- Reporting allocations with `-benchmem` and `b.ReportAllocs`
- Comparing implementations with sub-benchmarks
- Reading benchmark output
- Locking in allocation budgets with `testing.AllocsPerRun`

## Three implementations to compare

{{< example file="examples/m09/benchmarks/join.go" region="code" >}}

Before benchmarking anything, make sure the variants are **correct and equivalent**:

{{< example file="examples/m09/benchmarks/join_test.go" region="correctness" >}}

## Writing benchmarks

{{< example file="examples/m09/benchmarks/join_test.go" region="bench" run="go test -run xxx -bench BenchmarkJoin -benchmem ./examples/m09/benchmarks" >}}

`for b.Loop() { ... }` (Go 1.24) is the way to write the timed loop:

- The framework decides how many iterations are needed for a stable measurement.
- Code **before** the loop is setup and is excluded from the timing, so `b.ResetTimer()` is no longer needed.
- The compiler cannot optimize away calls inside the loop, so you do not need to assign results to a package-level "sink" variable.
- Since Go 1.26, `b.Loop` also no longer prevents inlining in the loop body.

Older code uses `for i := 0; i < b.N; i++`. It still works, but it is easy to get wrong (setup inside the loop, results optimized away). Prefer `b.Loop` in new code.

## Reading the output

Running the benchmark above on a 2-core cloud VM with Go 1.24 produced:

```text
BenchmarkJoin/plus/10-2           591.0 ns/op       296 B/op      18 allocs/op
BenchmarkJoin/builder/10-2         79.87 ns/op       64 B/op       1 allocs/op
BenchmarkJoin/append/10-2          95.57 ns/op       96 B/op       2 allocs/op
BenchmarkJoin/plus/1000-2      2002495 ns/op    5043136 B/op    2983 allocs/op
BenchmarkJoin/builder/1000-2     30777 ns/op       9984 B/op     986 allocs/op
BenchmarkJoin/append/1000-2      15133 ns/op      11008 B/op       2 allocs/op
```

| Column | Meaning |
|---|---|
| name and `-2` | benchmark and subtest name; the suffix is `GOMAXPROCS` |
| `ns/op` | average time per iteration |
| `B/op` | bytes allocated per iteration (`-benchmem` or `b.ReportAllocs`) |
| `allocs/op` | heap allocations per iteration |

Your numbers will differ, but the shape is the lesson:

- Concatenation with `+=` is **quadratic**: for 1000 numbers it allocated about 5 MB and was over 100 times slower, because every `+=` copies the whole string so far.
- `strings.Builder` was fastest for 10 numbers, but at 1000 it made 986 allocations. They come from `strconv.Itoa`, which allocates a new string for each number above 99 (smaller ones come from a cache).
- Appending digits directly with `strconv.AppendInt` made exactly **2 allocations** at any size: the buffer and the final string.

**Allocations often matter more than raw instruction counts**, because each one costs time now and garbage collection work later.

## Benchmark hygiene

- **Run several times** and compare distributions, not single numbers: `go test -bench . -count=10 > old.txt`, then compare with `benchstat` (lesson 10.1).
- **Close other programs**; laptops on battery or thermal throttling give noisy results.
- **Benchmark realistic inputs.** Results for 10 items and 1000 items told different stories above.
- `-benchtime=3s` or `-benchtime=1000x` controls duration or iteration count; `-cpu=1,2,4` repeats with different `GOMAXPROCS`.
- `b.RunParallel` benchmarks code under concurrent load (the rate limiter project has an example).

## Allocation budgets as tests

Benchmarks are run by hand. A **test** can lock in an allocation budget so a regression fails CI:

{{< example file="examples/m09/benchmarks/join_test.go" region="allocs" run="go test -run TestJoinAppendAllocations -v ./examples/m09/benchmarks" >}}

This test caught a real mistake while writing this lesson. The first version of `JoinAppend` sized its buffer at 4 bytes per number. With 4-digit numbers plus commas that was too small, the buffer grew once, and the test reported 3 allocations instead of 2. Allocation counts are exact, so these tests are not flaky the way timing assertions would be. (They can change between Go releases as the compiler improves, so keep budgets for code where it truly matters.)

## Common mistakes

### Benchmarking without checking correctness

A fast wrong answer is easy to produce. Test equivalence first.

### Setup inside the timed loop

Allocating the input on every iteration measures the setup, not the code. With `b.Loop`, put setup before the loop.

### Trusting one run

Variance of 5 to 10 percent between runs is normal. Use `-count` and `benchstat` before concluding anything.

### Asserting on timings in tests

`if elapsed > 10*time.Millisecond { t.Fail() }` is flaky on busy CI machines. Assert on allocations, or use benchmarks.

## Key takeaways

- Benchmarks are `BenchmarkXxx(b *testing.B)` functions; write the timed loop as `for b.Loop()`.
- `-benchmem` reports bytes and allocations per operation; allocations are often the real cost.
- Compare variants with sub-benchmarks on realistic input sizes.
- Run multiple times and compare with `benchstat`.
- Lock in allocation budgets with `testing.AllocsPerRun` tests.

## Further reading

- [Package testing: Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)
- [More predictable benchmarking with testing.B.Loop](https://go.dev/blog/testing-b-loop) on the Go blog
