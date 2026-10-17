---
title: "10.2 CPU and memory profiling with pprof"
linkTitle: "10.2 pprof"
weight: 2
description: "Collect CPU and memory profiles from benchmarks and running services, read top, list and flame graph views, and turn what the profile shows into a targeted fix."
goVersion: "1.27"
---

## What you will learn

- How sampling profilers work, and what CPU and heap profiles contain
- Collecting profiles from benchmarks with `-cpuprofile` and `-memprofile`
- Reading `top`, `top -cum` and `list` output
- The flame graph view in the web UI
- Profiling running services with `net/http/pprof`

## The idea

A **profiler** shows where a program spends its time or memory. Go's CPU profiler is a **sampling** profiler: about 100 times per second it records the stack of whatever is running. Functions that appear in many samples are where the time goes. The heap profiler similarly samples allocations, recording where memory was allocated.

Profiles answer "**where** should I optimize?". Benchmarks (lesson 10.1) answer "**did** it help?". You need both.

## Profiling a benchmark

```bash
go test -run xxx -bench 'BenchmarkNaive$' -benchtime=20x \
  -cpuprofile cpu.out -memprofile mem.out ./examples/m10/loganalyzer
go tool pprof -top cpu.out
```

`go test` keeps the compiled test binary (`loganalyzer.test`) next to the profiles so `pprof` can show function names and source lines.

## Reading `top`

Real output for the naive analyzer (trimmed):

```text
Type: cpu
Duration: 3.42s, Total samples = 3950ms (115.42%)
      flat  flat%   sum%        cum   cum%
     170ms  4.30%  4.30%      170ms  4.30%  runtime.nextFreeFast (inline)
     160ms  4.05%  8.35%      160ms  4.05%  runtime.futex
     160ms  4.05% 12.41%      160ms  4.05%  runtime.memclrNoHeapPointers
     140ms  3.54% 15.95%      300ms  7.59%  regexp.onePassCopy
     130ms  3.29% 19.24%      360ms  9.11%  regexp/syntax.(*compiler).inst
     120ms  3.04% 22.28%      570ms 14.43%  regexp.makeOnePass.func1
     120ms  3.04% 25.32%      710ms 17.97%  runtime.growslice
```

- **flat**: time spent in the function itself.
- **cum** (cumulative): time in the function **and everything it calls**.
- Total samples above 100% means more than one CPU was busy (the garbage collector runs in parallel).

The flat view is confusing here: the top entries are memory allocator and regexp internals, each a few percent. The **cumulative** view tells the real story:

```bash
go tool pprof -top -cum cpu.out
```

```text
      flat  flat%   sum%        cum   cum%
     0.01s  0.25%  0.25%      2.96s 74.94%  .../loganalyzer.AnalyzeNaive
         0     0%  0.25%      2.52s 63.80%  regexp.MustCompile
         0     0%  0.25%      2.52s 63.80%  regexp.compile
     0.06s  1.52%  1.77%      1.19s 30.13%  runtime.mallocgc
     0.04s  1.01%  2.78%      1.01s 25.57%  regexp.compileOnePass
```

**64% of all CPU time is `regexp.MustCompile`**, called from our function. And 30% is spent allocating memory.

## Reading `list`

`list` shows time per source line of a function:

```bash
go tool pprof -list AnalyzeNaive cpu.out
```

```text
      10ms      2.96s (flat, cum) 74.94% of Total
         .      2.52s     18:		re := regexp.MustCompile(`^(\S+) (\S+) (\S+) (\d{3}) (\d+)ms$`)
         .      330ms     19:		m := re.FindStringSubmatch(line)
         .       70ms     28:		key := fmt.Sprintf("%s", strings.ToLower(m[3]))
```

Line 18 is the culprit: the pattern is compiled again for **every line**. Matching (line 19) is cheap by comparison. The fix is to compile once, in a package-level variable, which is exactly `AnalyzeBetter`:

{{< example file="examples/m10/loganalyzer/better.go" region="better" >}}

Lesson 10.1's benchstat output confirms it: 92% faster.

## Memory profiles

```bash
go tool pprof -top -sample_index=alloc_space mem.out
```

```text
Type: alloc_space
      flat  flat%   sum%        cum   cum%
  499.47MB 22.26% 22.26%   499.47MB 22.26%  regexp/syntax.(*compiler).inst
  492.05MB 21.93% 44.20%   492.05MB 21.93%  regexp/syntax.(*parser).newRegexp
  400.77MB 17.86% 62.06%   400.77MB 17.86%  regexp.onePassCopy
```

The heap profile has four views, selected with `-sample_index`:

| Index | Shows |
|---|---|
| `alloc_space` | total bytes allocated over time (find **allocation churn**) |
| `alloc_objects` | number of allocations over time |
| `inuse_space` | bytes still live when the profile was taken (find **memory leaks** and bloat) |
| `inuse_objects` | live object count |

For speed, `alloc_space` and `alloc_objects` matter most: every allocation costs time now and garbage collection work later. For "why is memory growing", look at `inuse_space`.

## The second round

After fixing the regexp compile, profile `AnalyzeBetter` the same way and the remaining time is mostly regular expression matching and the allocations it makes for submatches. That led to `AnalyzeFast`, which scans with `strings.Cut` and avoids allocating keys that are already lowercase:

{{< example file="examples/m10/loganalyzer/fast.go" region="fast" >}}

This is the normal rhythm: **profile, fix the biggest thing, measure, repeat**. Stop when the code is fast enough for its purpose.

## The web UI and flame graphs

```bash
go tool pprof -http=:8080 cpu.out
```

This opens an interactive UI. Since Go 1.26 it opens on the **flame graph** view, where each bar is a function, its width is its share of samples, and bars stack on top of the functions that called them. Wide bars near the top of the stack are your hot spots. The UI also has graph, top, source and disassembly views.

## Profiling running services

Import `net/http/pprof` and serve it on an **internal** port:

```go
import _ "net/http/pprof" // registers /debug/pprof/ handlers on http.DefaultServeMux

go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()
```

```bash
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=30  # CPU
go tool pprof http://localhost:6060/debug/pprof/heap                              # memory
curl http://localhost:6060/debug/pprof/goroutine?debug=1                          # goroutines
```

Never expose these endpoints publicly: they reveal internals and can be used to slow the service down. Profiling in production is cheap enough to do routinely; continuous profiling services collect these profiles from every instance, which is also how you get profiles for PGO (lesson 10.5).

## Common mistakes

### Optimizing by flat time only

Runtime functions (`mallocgc`, `memclr`) often top the flat list. Use `-cum` and `list` to find **your** code that causes them.

### Profiling a too-short run

A profile with a few dozen samples is noise. Run long enough to get thousands (use `-benchtime`).

### Forgetting the binary

Without the binary, `pprof` cannot always resolve symbols for older profiles. Keep the test binary or build with the same source.

## Key takeaways

- CPU profiles sample stacks; heap profiles sample allocations.
- Collect from benchmarks with `-cpuprofile` and `-memprofile`, or from services with `net/http/pprof`.
- `top -cum` and `list` lead you from runtime noise to your own code.
- Use `alloc_space` for churn, `inuse_space` for leaks.
- Profile, fix the biggest problem, measure with benchstat, repeat.

## Further reading

- [Profiling Go Programs](https://go.dev/blog/pprof) on the Go blog
- [Diagnostics](https://go.dev/doc/diagnostics) on go.dev
- [Package net/http/pprof](https://pkg.go.dev/net/http/pprof)
