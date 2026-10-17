---
title: "10. Profiling and performance"
weight: 10
description: "Measure with benchmarks and benchstat, find hot spots with pprof, see scheduling with the execution tracer and flight recorder, cut allocations with escape analysis, apply PGO and sync.Pool."
---

This module follows one rule: **measure first**. Every claim about speed in it comes from a benchmark or a profile that was actually run, on a 2-core cloud VM, and you can rerun every one of them yourself. The running example is a log analyzer that gets about 140 times faster over the course of the module, guided by data at every step.

**Time:** about 3 to 4 hours, plus the exercise.

{{< cards >}}
  {{< card link="01-measure-first" title="10.1 Measure first: benchstat" subtitle="Benchmarks you can trust and compare." >}}
  {{< card link="02-pprof" title="10.2 CPU and memory profiling" subtitle="Find where time and memory really go." >}}
  {{< card link="03-tracer" title="10.3 The execution tracer" subtitle="Latency, scheduling, and the flight recorder." >}}
  {{< card link="04-allocations" title="10.4 Allocations and escape analysis" subtitle="Why values move to the heap, and how to stop it." >}}
  {{< card link="05-pgo" title="10.5 Profile-guided optimization" subtitle="Free speed from production profiles." >}}
  {{< card link="06-sync-pool" title="10.6 sync.Pool" subtitle="Reusing objects, and when not to." >}}
  {{< card link="07-exercise" title="10.7 Exercise: optimize under a budget" subtitle="Make correct code fast, measured." >}}
{{< /cards >}}
