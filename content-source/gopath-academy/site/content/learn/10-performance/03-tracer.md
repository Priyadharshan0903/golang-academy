---
title: "10.3 The execution tracer and flight recorder"
linkTitle: "10.3 Execution tracer"
weight: 3
description: "Use the execution tracer to see goroutine scheduling, blocking, GC pauses and latency, annotate traces with tasks and regions, and capture traces of rare events with the flight recorder."
goVersion: "1.27"
---

## What you will learn

- What the execution tracer shows that profiles cannot
- Collecting traces from tests and programs, and opening them with `go tool trace`
- Annotating traces with tasks, regions and logs
- Capturing the moments before a rare problem with the **flight recorder** (Go 1.25)

## Profiles vs traces

A CPU profile tells you **where** CPU time goes, summed over the whole run. It cannot tell you why a particular request took 2 seconds while using only 5 milliseconds of CPU. That request was **waiting**: for a lock, a channel, the network, the garbage collector or a free processor.

The **execution tracer** records events with timestamps: goroutines starting, blocking and unblocking, system calls, GC phases, processor (P) activity. It answers questions about **latency and concurrency**:

- Why is this request slow when CPU usage is low?
- Are my workers actually running in parallel, or waiting on one lock?
- How long do GC pauses and assists take, and when?

## Collecting a trace

```bash
# From tests or benchmarks:
go test -run TestSomething -trace trace.out ./pkg
go tool trace trace.out
```

In a program:

```go
f, _ := os.Create("trace.out")
trace.Start(f)
defer trace.Stop()
```

Or from a running service with `net/http/pprof`: `curl -o trace.out http://localhost:6060/debug/pprof/trace?seconds=5`.

`go tool trace` opens a browser UI with a timeline per processor and per goroutine, views of goroutine analysis, network and synchronization blocking, and scheduler latency. Tracing has a modest overhead (a few percent with the current implementation), so it is usable in production for short periods.

## Annotating traces

```go
ctx, task := trace.NewTask(ctx, "handleOrder")   // a logical operation, possibly across goroutines
defer task.End()

trace.WithRegion(ctx, "validate", func() { validate(order) })  // a timed region in one goroutine
trace.Log(ctx, "orderID", order.ID)                            // a message attached to the task
```

Tasks and regions appear in the trace UI, with per-task latency histograms. This turns "the trace of everything" into "the trace of slow `handleOrder` calls".

## The flight recorder (Go 1.25)

The hardest performance problems are rare: one request in ten thousand takes 5 seconds. You cannot trace everything all the time, and by the time you notice, the moment has passed.

The **flight recorder**, added in Go 1.25, keeps the last few seconds of trace data in a ring buffer in memory. When your code notices something interesting, it writes a snapshot of what led up to it:

{{< example file="examples/m10/flightrec/flightrec.go" region="recorder" >}}

Use it with a latency threshold in a request middleware, or on errors that are hard to reproduce. Write the snapshot to a file (or upload it), then open it with `go tool trace` like any other trace. The `MinAge` and `MaxBytes` settings bound how much history is kept.

## Common mistakes

### Using CPU profiles for latency problems

If CPU is low but latency is high, the program is waiting. Trace it.

### Tracing for too long

Traces are large. A few seconds is usually enough; use the flight recorder for rare events.

### Unannotated traces of busy services

Thousands of goroutines are hard to read. Add tasks and regions around the operations you care about.

## Key takeaways

- The tracer records timestamped scheduling, blocking, syscall and GC events; use it for latency and concurrency questions.
- Collect with `go test -trace`, `trace.Start`, or `/debug/pprof/trace`, and view with `go tool trace`.
- Annotate with `trace.NewTask`, `trace.WithRegion` and `trace.Log`.
- The flight recorder (Go 1.25+) keeps recent trace data in memory and snapshots it when something goes wrong.

## Further reading

- [Package runtime/trace](https://pkg.go.dev/runtime/trace)
- [More powerful Go execution traces](https://go.dev/blog/execution-traces-2024) on the Go blog
- [Flight Recorder in Go 1.25](https://go.dev/blog/flight-recorder) on the Go blog
