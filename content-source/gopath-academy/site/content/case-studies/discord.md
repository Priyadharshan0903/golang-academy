---
title: "Discord: when Go's garbage collector was the wrong fit"
linkTitle: "Discord: Go to Rust"
weight: 8
description: "Discord's Read States service, a huge in-memory LRU cache written in Go 1.9, showed latency spikes every two minutes from forced garbage collections scanning the whole cache. They rewrote it in Rust. A fair look at what happened, and what has and has not changed."
---

| | |
|---|---|
| **Company** | Discord |
| **System** | Read States: which channels and messages each user has read |
| **Go version** | 1.9.2 (1.8, 1.9 and 1.10 also tried) |
| **Source** | Jesse Howarth, [Why Discord is switching from Go to Rust](https://discord.com/blog/why-discord-is-switching-from-go-to-rust), Discord blog, 4 February 2020 |
| **Connects to** | Lesson 1.1 (when not to use Go), lesson 11.3 |

## The story

Every Discord user has a "read state" per channel: how far they have read, how many mentions they have. The **Read States** service sits on the hot path, used when users connect, send and read messages. Each server kept an **LRU cache** with millions of users and tens of millions of read states in memory, handling hundreds of thousands of cache updates per second, backed by Cassandra.

The Go version was fast on average, but latency and CPU **spiked about every two minutes**. The cause: if no garbage collection has run for two minutes, the Go runtime **forces one**. The cache allocated little new memory, so collections rarely triggered on their own, and when the forced collection came it had to **scan the entire LRU cache**, tens of millions of objects full of pointers.

They tried the obvious tuning:

- **Changing `GOGC` on the fly** did nothing: the problem was not collections being triggered too often, but the cost of the forced one.
- **Smaller caches**, partitioned into many per server, reduced the spikes, but raised p99 latency because more lookups went to the database.

In May 2019 they finished a **Rust** rewrite. With no garbage collector, the spikes disappeared; after optimization (including a `BTreeMap` instead of a `HashMap` for the cache, and fewer copies) the Rust version beat the hand-tuned Go version on latency, CPU and memory, and they raised the cache to 8 million read states per server. The post shows the before and after comparisons as graphs, without exact figures in the text.

## A fair reading

This is the best-known story of a team leaving Go, and it is worth reading precisely:

- **It is about one service with one memory shape**: a huge, long-lived, pointer-heavy in-memory cache. A collector that must periodically trace the live heap pays for exactly that shape. Go's GC cost scales with the live heap, not with the garbage (lesson 11.3).
- **The forced collection is still there.** Go still runs a collection if none has run for two minutes (`forcegcperiod` in the runtime). `GOMEMLIMIT`, added in Go 1.19, would not have removed that scan. The Green Tea collector in Go 1.26 makes marking cheaper per object, which helps, but a forced scan of tens of millions of objects remains a scan.
- **Go-side mitigations exist**, at a cost: storing the cache in pointer-free structures (for example, values in large slices indexed by integers, which the GC does not need to trace), or moving it off the Go heap. These make Go code look much less like Go.
- **Discord did not leave Go everywhere.** The post is about Read States; the author describes Rust adoption across Discord's stack, not a ban on Go.

## What you can apply

- **Know your memory shape.** Large, long-lived, pointer-heavy heaps are the worst case for a tracing GC. Measure live heap and GC CPU before choosing a design (lessons 10.2 and 11.3).
- **Pointer-free data is cheap for the GC.** `[]byte`, arrays of plain structs and integer indexes instead of pointers can shrink GC work dramatically (lesson 11.3).
- **Choosing a different language for one component is a legitimate engineering decision.** Lesson 1.1 lists when Go is not the right tool; this is a textbook example.

## Sources

- Jesse Howarth: [Why Discord is switching from Go to Rust](https://discord.com/blog/why-discord-is-switching-from-go-to-rust), Discord blog, 4 February 2020
- The Go runtime source: `forcegcperiod`, the two-minute forced collection, in [`runtime/proc.go`](https://github.com/golang/go/blob/master/src/runtime/proc.go)
