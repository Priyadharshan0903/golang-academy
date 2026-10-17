---
title: "7.2 Worker pools"
linkTitle: "7.2 Worker pools"
weight: 2
description: "Process a stream of jobs with a fixed number of goroutines: the four parts of a worker pool, collecting results in order, cancellation, and sizing the pool."
goVersion: "1.27"
---

## What you will learn

- Why bounding concurrency matters
- The four parts of a worker pool: workers, feeder, closer and collector
- How to keep results in input order
- How to make every part of the pool cancellable
- How to choose the number of workers

## The idea

"One goroutine per item" is fine for ten items. For a million it is not: you open a million connections, allocate a million stacks, and overwhelm whatever you are calling. A **worker pool** runs a **fixed number** of goroutines that pull jobs from a channel, so concurrency stays bounded no matter how much work arrives.

## The pattern

{{< example file="examples/m07/workerpool/workerpool_test.go" region="pool" run="go test -race -run Example_workerPool -v ./examples/m07/workerpool" >}}

Four pieces, each with one job:

1. **Workers** (`workers` goroutines) loop over `jobs` until it is closed. They are the only goroutines doing real work.
2. **The feeder** sends every job, then **closes** `jobs`, which ends the workers' `range` loops.
3. **The closer** waits for all workers (`wg.Wait()`), then closes `results`. This must be a separate goroutine: if the collector called `Wait` itself, it would deadlock, because workers block on sending results that nobody is receiving yet.
4. **The collector** (here, the calling function) ranges over `results` until it is closed.

Every send selects on `ctx.Done()` too, so cancelling the context unblocks every goroutine and nothing leaks.

## Keeping results in order

Workers finish in any order. Carry the input's **index** with each job and result, and write results into a pre-sized slice by index, as `counts[r.id] = r.words` does. This is cheaper and simpler than sorting afterwards.

## Handling errors

The example has no errors. With real work you choose a policy:

- **Fail fast**: stop everything on the first error. Cancel a shared context and return the error. `errgroup` (lesson 7.4) packages exactly this.
- **Collect all**: keep going, put each error in its result, and report them together (for example with `errors.Join`, lesson 5.4).

The `pmap` exercise below implements fail fast.

## How many workers?

| Work type | Starting point |
|---|---|
| CPU-bound (hashing, parsing, compression) | `runtime.GOMAXPROCS(0)`, the number of usable CPUs |
| I/O-bound (HTTP calls, database queries) | Higher, limited by what the **other side** can handle: connection pool size, API rate limits |
| Mixed | Measure. Start small and increase while throughput improves |

More workers than useful only adds contention. For I/O, the downstream service's limits usually decide the number, which is why worker pools pair naturally with rate limiting (lesson 7.5).

## Common mistakes

### Calling `wg.Wait()` before reading results

Workers block sending results, `Wait` never returns: deadlock. Close `results` from a separate goroutine.

### Forgetting to close the jobs channel

Workers wait forever for the next job. The feeder must `close(jobs)` when done.

### A worker that returns early without draining

If one worker exits on error while the feeder keeps sending, the remaining workers handle the load, but if **all** exit, the feeder blocks forever. Make the feeder select on `ctx.Done()` and cancel the context when workers give up.

### Unbounded result buffering

Making `results` a channel with a huge buffer to "avoid blocking" hides backpressure and can use lots of memory. Let the collector set the pace.

## Exercise

{{< exercise dir="m07/pmap" >}}
Write a generic `Map` that runs a function over a slice with a fixed number of workers, returns results **in input order**, cancels the remaining work on the first error, and never leaks goroutines. The tests check ordering, the concurrency limit, cancellation and leaks.
{{< /exercise >}}

## Key takeaways

- A worker pool bounds concurrency with a fixed number of goroutines reading from a jobs channel.
- Four roles: workers, a feeder that closes jobs, a closer that waits and closes results, and a collector.
- Carry indexes to keep results in order.
- Make every send cancellable with `select` on `ctx.Done()`.
- Size CPU-bound pools by CPUs, I/O-bound pools by downstream capacity.

## Further reading

- [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines) on the Go blog
- [Package golang.org/x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup), the fail-fast pool in a box (lesson 7.4)
