---
title: "10.5 Profile-guided optimization (PGO)"
linkTitle: "10.5 PGO"
weight: 5
description: "Feed a CPU profile from production into the compiler so it inlines and devirtualizes the hot paths, with a real measurement on the log analyzer."
goVersion: "1.27"
---

## What you will learn

- What profile-guided optimization is and what the compiler does with a profile
- The `default.pgo` workflow for real services
- A measured result on this module's log analyzer
- What PGO does not do

## The idea

The compiler normally optimizes without knowing which code runs most. With **profile-guided optimization** (stable since Go 1.21), you give it a CPU profile from a representative run, and it spends its optimization budget where the profile says time is spent. Mainly it:

- **inlines** hot functions more aggressively, even ones over the normal size limit;
- **devirtualizes** hot interface method calls: if the profile shows a call through an interface almost always reaches one concrete type, the compiler adds a fast direct call for that type (lesson 11.5 explains why that helps).

You change no code. The Go team reports typical improvements of a few percent up to around 14% for representative programs.

## The workflow for a service

1. **Collect a CPU profile from production**, or from a load test that resembles it. With `net/http/pprof`: `curl -o cpu.pprof http://localhost:6060/debug/pprof/profile?seconds=30`. Merging profiles from several instances (`go tool pprof -proto a.pprof b.pprof > merged.pprof`) gives a more representative result.
2. **Save it as `default.pgo`** in the directory of your `main` package, and commit it.
3. **Build normally.** `go build` uses `default.pgo` automatically (the default is `-pgo=auto`). Use `-pgo=off` to disable it, or `-pgo=path/to/file` to choose a profile.
4. **Refresh the profile** from time to time, for example every release. A slightly stale profile still helps; Go's PGO is designed to tolerate code changes.

## A measurement

For the log analyzer, a profile was collected by running the `better` and `fast` benchmarks, and the benchmarks were then run 8 times without and with that profile (`go test -pgo=cpu.pprof`). Result from the 2-core test VM with Go 1.24:

```text
                      │  without PGO  │            with PGO                 │
                      │    sec/op     │    sec/op     vs base               │
Analyze/impl=better-2    10.88m ±  7%   10.06m ± 10%   -7.56% (p=0.021 n=8)
Analyze/impl=fast-2      1026.0µ ± 5%   935.0µ ±  6%   -8.88% (p=0.000 n=8)
geomean                   3.342m         3.067m        -8.22%
```

About **8% faster for free**, statistically significant even on a noisy VM. Your results will vary with the code: programs with many small hot functions and interface calls benefit most.

To reproduce:

```bash
cd examples/m10/loganalyzer
go test -run xxx -bench 'Analyze/impl=(better|fast)' -benchtime=2s -cpuprofile cpu.pprof .
go test -run xxx -bench 'Analyze/impl=(better|fast)' -count=8 . > nopgo.txt
go test -pgo=cpu.pprof -run xxx -bench 'Analyze/impl=(better|fast)' -count=8 . > pgo.txt
benchstat nopgo.txt pgo.txt
```

## What PGO does not do

- It does not fix algorithmic problems. Compiling a regexp per line is still slow with PGO.
- It does not reduce allocations directly (though inlining sometimes lets escape analysis keep values on the stack).
- Builds take somewhat longer, and a profile from an unrepresentative workload can optimize the wrong paths.

Do the big, obvious fixes first (lessons 10.2 and 10.4), then add PGO as a final, low-effort gain.

## Key takeaways

- PGO feeds a CPU profile to the compiler, which inlines and devirtualizes hot code.
- Put a production-representative profile in `default.pgo` next to your `main` package; `go build` uses it automatically.
- Expect a few percent to around 14%; this module's example measured about 8%.
- PGO complements, and never replaces, profiling and fixing your code.

## Further reading

- [Profile-guided optimization](https://go.dev/doc/pgo) on go.dev
- [Profile-guided optimization in Go 1.21](https://go.dev/blog/pgo) on the Go blog
