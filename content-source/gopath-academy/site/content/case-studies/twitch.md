---
title: "Twitch: a 10 GiB memory ballast, and why GOMEMLIMIT replaced it"
linkTitle: "Twitch: GC tuning"
weight: 7
description: "Twitch's API gateway ran 8 to 10 GC cycles per second with a small heap on a 64 GiB machine. A large, untouched byte array cut GC cycles by 99% and p99 latency by 45%. Go 1.19's GOMEMLIMIT now does the same job properly."
---

| | |
|---|---|
| **Company** | Twitch |
| **System** | Visage, the API front end and gateway for external API traffic |
| **Go version** | 1.11 |
| **Source** | Ross Engers, [Go memory ballast: How I learnt to stop worrying and love the heap](https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap/), Twitch blog, 10 April 2019 |
| **Connects to** | Lesson 11.3, Module 10 |

## The story

Visage is the gateway that all of Twitch's external API traffic passes through, running on EC2 machines with **64 GiB** of memory. Its live heap was small, under **450 MiB** on average. Because the default `GOGC=100` triggers a collection whenever the heap doubles (lesson 11.3), a small live heap and a high allocation rate meant very frequent collections: about **8 to 10 GC cycles per second** at steady state, with **over 30% of CPU** spent in GC-related functions, and more during "refresh storms" when traffic could reach about 20 times normal.

The fix was a **memory ballast**: allocate one huge byte slice at startup and keep it alive.

```go
ballast := make([]byte, 10<<30) // 10 GiB, as in the article's example
```

The GC counts the ballast as live heap, so the next collection is not triggered until the heap grows by another 10 GiB. Because the slice is never written, the operating system never backs it with physical memory: it costs address space, not RAM. And because it is a `[]byte` with no pointers, marking it costs almost nothing.

The article does not state the production size, and advises starting small and increasing with testing. The reported results:

- **GC cycles down about 99%**;
- **CPU per machine down about 30%**;
- **p99 API latency at peak down about 45%**.

GC **pauses** did not change much. The latency gain came from **mark assists**: when the GC runs, goroutines that allocate are drafted to help with marking (lesson 11.3), and those goroutines were serving requests. Fewer GC cycles meant fewer requests slowed down by assist work.

The author also explains why they did not just raise `GOGC`: the amount of memory was what mattered, not the ratio, and a very high `GOGC` would make the heap size depend sharply on small changes in live memory.

## What changed since

The ballast was a clever workaround for a missing feature, and **Go 1.19 added that feature: `GOMEMLIMIT`** (lesson 11.3). The modern equivalent of Twitch's fix is:

```bash
GOGC=off GOMEMLIMIT=8GiB ./visage   # collect only as memory approaches 8 GiB
```

or, more conservatively, keep `GOGC` at its default and set `GOMEMLIMIT` to protect against running out of memory. Lesson 11.3's experiment shows the same effect the ballast had: with `GOGC=off` and a memory limit, the GC runs only near the limit, far less often than with a small heap and the default `GOGC`. The Go team's [GC guide](https://go.dev/doc/gc-guide) discusses this use of the memory limit directly. Since Go 1.26, the Green Tea collector also reduces the CPU cost of each cycle for programs like this.

## What you can apply

- **Look at GC frequency, not just pause time.** `GODEBUG=gctrace=1` or the runtime metrics show cycles per second (lesson 11.3); assists hurt latency even when pauses are tiny.
- **A small live heap with high allocation churn** is the classic case for raising the GC's target, today with `GOMEMLIMIT` rather than a ballast.
- **Reduce allocations first** (lesson 10.4): fewer allocations mean fewer cycles, whatever the settings.
- **Measure tail latency** (p99), where GC assists show up.

## Sources

- Ross Engers: [Go memory ballast: How I learnt to stop worrying and love the heap](https://blog.twitch.tv/en/2019/04/10/go-memory-ballast-how-i-learnt-to-stop-worrying-and-love-the-heap/), Twitch blog, 10 April 2019
- The Go team: [A Guide to the Go Garbage Collector](https://go.dev/doc/gc-guide)
