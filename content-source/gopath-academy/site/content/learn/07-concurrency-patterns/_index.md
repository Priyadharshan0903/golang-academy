---
title: "7. Concurrency patterns and context"
weight: 7
description: "context for cancellation and deadlines, worker pools, pipelines, errgroup, rate limiting, goroutine leaks and graceful shutdown, plus two projects."
---

Module 6 gave you the primitives. This module turns them into the patterns that real Go services are made of, all built around `context.Context`, the standard way to carry cancellation and deadlines through a program.

**Time:** about 4 to 5 hours, plus exercises and two projects. Keep using `-race`.

{{< cards >}}
  {{< card link="01-context" title="7.1 context.Context" subtitle="Cancellation, deadlines, causes and values." >}}
  {{< card link="02-worker-pools" title="7.2 Worker pools" subtitle="A fixed number of workers over a stream of jobs." >}}
  {{< card link="03-pipelines" title="7.3 Pipelines and fan-out/fan-in" subtitle="Composable, cancellable stages." >}}
  {{< card link="04-errgroup" title="7.4 errgroup and bounded concurrency" subtitle="First error cancels the rest." >}}
  {{< card link="05-rate-limiting" title="7.5 Rate limiting and backpressure" subtitle="Semaphores, tickers, token buckets." >}}
  {{< card link="06-goroutine-leaks" title="7.6 Goroutine leaks" subtitle="Causes, detection and the goroutineleak profile." >}}
  {{< card link="07-graceful-shutdown" title="7.7 Graceful shutdown" subtitle="Signals, draining and deadlines." >}}
  {{< card link="08-projects" title="7.8 Projects: rate limiter and crawler" subtitle="Two complete concurrent programs." >}}
{{< /cards >}}
