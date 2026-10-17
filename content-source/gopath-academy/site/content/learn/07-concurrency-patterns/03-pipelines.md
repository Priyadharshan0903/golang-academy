---
title: "7.3 Pipelines and fan-out / fan-in"
linkTitle: "7.3 Pipelines"
weight: 3
description: "Compose concurrent stages connected by channels, spread work across stages with fan-out, merge results with fan-in, and stop the whole pipeline early with context."
goVersion: "1.27"
---

## What you will learn

- How to structure work as a pipeline of stages connected by channels
- Fan-out: several goroutines reading from one channel
- Fan-in: merging several channels into one
- How to stop every stage when the consumer is done early
- When a pipeline is the wrong shape for a problem

## The idea

A **pipeline** is a series of **stages** connected by channels. Each stage is a function that receives values from an inbound channel, does something, and sends values on an outbound channel. Stages run concurrently, so while stage 3 processes item 1, stage 2 can work on item 2 and stage 1 on item 3.

You built a pipeline without cancellation in the lesson 6.2 exercise. Here every stage also takes a `ctx`.

## Cancellable stages

{{< example file="examples/m07/pipelines/pipelines_test.go" region="stages" >}}

Every stage follows the same template:

1. Create the outbound channel.
2. Start a goroutine that `defer close(out)` and loops over the input.
3. Every send is a `select` with `case <-ctx.Done(): return`.
4. Return the outbound channel as receive-only.

With that template, any stage can be combined with any other.

## Fan-out and fan-in

{{< example file="examples/m07/pipelines/pipelines_test.go" region="fanout" run="go test -race -run Example_fanOutFanIn -v ./examples/m07/pipelines" >}}

- **Fan-out**: several `square` stages read from the **same** source channel. Each value is received by exactly one of them, so the work is shared. Use it when one stage is slower than the others.
- **Fan-in**: `merge` combines several channels into one, closing the output when all inputs are closed.

Fan-out loses ordering. If order matters, carry an index and reorder at the end, as in the worker pool lesson.

## Stopping early

The consumer might not want every value. Without cancellation, upstream stages would block forever trying to send values nobody reads:

{{< example file="examples/m07/pipelines/pipelines_test.go" region="early" run="go test -race -run Example_stopEarly -v ./examples/m07/pipelines" >}}

`defer cancel()` in the consumer closes `ctx.Done()` when it returns, every stage's send `select` picks the `Done` case, the stages return, and their deferred `close` calls cascade down the pipeline. No goroutine is left behind.

## Pipelines vs iterators vs plain loops

Not every sequence of steps needs goroutines:

| Use | When |
|---|---|
| A plain loop | Each step is cheap; there is nothing to overlap |
| Iterators (lesson 3.8) | You want lazy, composable stages but no concurrency |
| A channel pipeline | Stages do slow work (I/O) that can overlap, or a stage benefits from fan-out |

Channel operations cost far more than function calls. A pipeline of trivial CPU stages is usually **slower** than a single loop. Measure (Module 10).

## Common mistakes

### Stages without cancellation

The consumer stops reading, upstream goroutines block on send forever. Every send needs `case <-ctx.Done()`.

### Closing a channel from the wrong place

Only the stage that owns the outbound channel closes it, and `merge` closes only after all its inputs are done.

### Pipelines for CPU-trivial work

Splitting `x * x` into its own goroutine costs more than it saves.

## Key takeaways

- A pipeline is a chain of stages connected by channels, each running concurrently.
- Each stage owns and closes its outbound channel and selects on `ctx.Done()` for every send.
- Fan-out shares work among several goroutines; fan-in merges their results.
- `defer cancel()` in the consumer tears the whole pipeline down cleanly.
- Use pipelines for overlapping slow work; use loops or iterators otherwise.

## Further reading

- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines) on the Go blog
- [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), Rob Pike's talk
