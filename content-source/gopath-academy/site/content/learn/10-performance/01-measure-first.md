---
title: "10.1 Measure first: benchmarks and benchstat"
linkTitle: "10.1 Measure first"
weight: 1
description: "Set up a fair comparison between implementations, run benchmarks repeatedly, and use benchstat to tell real improvements from noise."
goVersion: "1.27"
---

## What you will learn

- Why intuition about performance is unreliable, and what to do instead
- How to structure a fair comparison: equivalence test first, then benchmarks
- How to use `benchstat` to compare runs with statistics
- How to read `±` and `p=` in benchstat output

## The idea

Performance work without measurement is guessing. Experienced engineers guess wrong all the time, because the cost of code depends on the compiler, the runtime, the CPU cache and the input data. The workflow for any optimization is:

1. **Write a test** proving the old and new code produce the same results.
2. **Write a benchmark** with realistic input.
3. **Run it several times** for each version.
4. **Compare statistically** with `benchstat`.
5. Only then **profile** to find the next thing to fix (lesson 10.2).

## The running example

A log analyzer that summarizes lines like `2026-10-16T09:30:00Z GET /api/users 200 12ms`. Here is the first version, written for clarity:

{{< example file="examples/m10/loganalyzer/naive.go" region="naive" >}}

Two later versions, which the next lessons motivate, are `AnalyzeBetter` and `AnalyzeFast`. First, an equivalence test so we can never "optimize" into a wrong answer:

{{< example file="examples/m10/loganalyzer/analyzer_test.go" region="equivalence" run="go test -run TestImplementationsAgree -v ./examples/m10/loganalyzer" >}}

Note the extra malformed lines at the end. Optimized parsers are where edge cases break, so the test feeds them deliberately.

## One benchmark, several implementations

{{< example file="examples/m10/loganalyzer/analyzer_test.go" region="bench" >}}

Naming sub-benchmarks `impl=naive`, `impl=better` and `impl=fast` lets benchstat put them side by side as columns.

## benchstat

`benchstat` is the Go team's tool for comparing benchmark results. Install it once:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

Then run each benchmark several times (6 to 10 is typical) and compare:

```bash
go test -run xxx -bench BenchmarkAnalyze -benchmem -count=6 ./examples/m10/loganalyzer > bench.txt
benchstat -col /impl bench.txt
```

Actual output from the 2-core test VM, Go 1.24 (10,000 lines per operation):

```text
          │     naive      │                better                 │                fast                 │
          │     sec/op     │     sec/op      vs base               │    sec/op     vs base               │
Analyze-2   139186.4µ ± 9%   10538.1µ ± 16%  -92.43% (p=0.002 n=6)   997.4µ ± 15%  -99.28% (p=0.002 n=6)

          │      naive       │                better                 │                fast                 │
          │       B/op       │      B/op       vs base               │     B/op      vs base               │
Analyze-2   115223.15Ki ± 0%   1881.64Ki ± 0%  -98.37% (p=0.002 n=6)   29.15Ki ± 0%  -99.97% (p=0.002 n=6)

          │     naive      │               better                │                fast                │
          │   allocs/op    │  allocs/op    vs base               │  allocs/op   vs base               │
Analyze-2   1401.426k ± 0%   21.429k ± 0%  -98.47% (p=0.002 n=6)   1.638k ± 0%  -99.88% (p=0.002 n=6)
```

How to read it:

- Each cell is the **median** of the runs, and `±` is the confidence interval. `± 16%` says this VM is noisy; that is why one run is never enough.
- `vs base` is the change relative to the first column.
- `p=0.002` is the probability of seeing a difference this large if there were **no real difference**. benchstat reports `~` instead of a percentage when the difference is not statistically significant (by default p ≥ 0.05). When you see `~`, you have not proven anything.
- `n=6` is the number of runs used.

So: compiling the regular expression once (`better`) made the analysis about **13 times faster** and cut allocations by 98%. Replacing the regular expression with direct string scanning (`fast`) made it about **140 times faster** than the original. The next lesson shows how the profile pointed at exactly these changes.

### Comparing before and after a change

The more common workflow compares two versions of the **same** code:

```bash
go test -run xxx -bench . -count=10 ./pkg > old.txt
# ... make your change ...
go test -run xxx -bench . -count=10 ./pkg > new.txt
benchstat old.txt new.txt
```

## Making benchmarks trustworthy

- **Realistic input.** The analyzer benchmark generates 10,000 varied lines with a fixed random seed, including malformed ones. Benchmarking one repeated line would let caches and branch predictors flatter the code.
- **Stable machine.** Close other programs. On laptops, plug in power. In CI, results from shared runners are noisy; use them for big regressions only.
- **Enough runs.** If `±` is large, increase `-count`.
- **Look at allocations too.** `-benchmem` often explains a time difference.

## Common mistakes

### Concluding from one run

Run-to-run noise of 10% or more is normal on shared machines. Use `-count` and benchstat.

### Benchmarking the wrong thing

Tiny inputs that fit in cache, or setup code inside the loop. Check what the loop actually measures.

### Optimizing without a reason

Only optimize code that a profile or a production metric shows matters. A 99% speedup of something that takes 0.1% of total time saves nothing.

## Key takeaways

- Test equivalence first, then benchmark realistic inputs.
- Run benchmarks repeatedly and compare with `benchstat`; trust only significant results (no `~`).
- Name sub-benchmarks `key=value` to compare implementations as columns with `-col`.
- Look at `B/op` and `allocs/op` alongside time.

## Further reading

- [benchstat documentation](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
- [Package testing: Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)
