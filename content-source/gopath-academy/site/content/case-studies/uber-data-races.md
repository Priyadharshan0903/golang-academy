---
title: "Uber: what the race detector found in 46 million lines of Go"
linkTitle: "Uber: 2,000 data races"
weight: 6
description: "Uber ran Go's race detector across its 46-million-line monorepo and found over 2,000 data races. Their PLDI 2022 paper catalogs the Go-specific patterns behind them: loop variable capture, err capture, maps, slices, copied mutexes and WaitGroup misuse."
---

| | |
|---|---|
| **Company** | Uber |
| **Scale** | About 46 million lines of Go and 2,100 microservices in one monorepo |
| **Source** | Milind Chabbi, Murali Krishna Ramanathan, [A Study of Real-World Data Races in Golang](https://arxiv.org/abs/2204.00764), PLDI 2022 |
| **Connects to** | Module 6 (especially 6.6), Module 9 |

## The story

Uber's Go monorepo held about **46 million lines of Go** and **2,100 microservices** (for comparison, its Java monorepo had 19 million lines). In April 2021 the team began running Go's race detector (lesson 6.6) over it systematically. Not in CI at first: running the 100,000+ unit tests with `-race` raised the 95th percentile test time from about 25 to about 100 minutes. Instead they ran it **daily**, de-duplicated the reports, and filed each race as a task for its owners.

In about six months:

- **over 2,000 data races** were detected;
- **1,011 were fixed**, by **210 engineers**, in 790 unique patches;
- after the backlog, about **five new races per day** kept appearing.

### Why Go code has so many races

The paper measures how much more concurrency Go code uses than Java code at Uber: the median Go process ran **2,048 goroutines** against 256 threads for Java, and Go code used about 3.7 times more point-to-point synchronization constructs per million lines. More concurrency means more chances to get it wrong. The race patterns it catalogs are worth memorizing, because most are specific to Go:

| Pattern | Example |
|---|---|
| **Loop variable capture** | goroutines started in a loop all reading the same loop variable |
| **`err` capture** | a goroutine assigning the enclosing function's `err` |
| **Named results and `defer`** | deferred closures writing named results concurrently with the function |
| **Slices** | copying a slice header outside the lock that protects `append` |
| **Maps** | assuming writes to *different keys* of a map are safe; they are not |
| **Copied mutexes** | passing a struct containing a `sync.Mutex` by value, so each copy has its own lock |
| **Mixing channels and shared memory** | a future that writes a field and also signals on a channel, read without synchronization |
| **`WaitGroup.Add` inside the goroutine** | `Wait` can return before the goroutine even registers |
| **Parallel subtests** | `t.Parallel()` table tests sharing state |
| **Writes under `RLock`** | taking the read lock, then modifying |

## What changed since

Two of the most common patterns have since been addressed in the language and library:

- **Go 1.22 made loop variables per-iteration**, so goroutines started in a loop each capture their own variable (lesson 2.5). This removed the first row of the table for modules that declare `go 1.22` or later.
- **`sync.WaitGroup.Go` (Go 1.25)** calls `Add` before starting the goroutine, making the "`Add` inside the goroutine" mistake impossible when you use it (lesson 6.4).

`go vet`'s `copylocks` check catches copied mutexes, and the remaining patterns (maps, slices, captured `err`, mixed channels and memory) are exactly why the race detector exists.

## What you can apply

- **Run tests with `-race`**, in CI if you can; Uber's numbers show what accumulates when you do not (lesson 6.6).
- **Never assume map writes to different keys are safe.** Protect the whole map (lesson 6.4).
- **Prefer `wg.Go`**, and declare `go 1.22` or later in `go.mod` to get per-iteration loop variables.
- **Do not copy structs containing locks**: pass pointers, and let `go vet` check.
- **Pick one coordination style per piece of state**: channels or locks, not both.

## Sources

- Milind Chabbi, Murali Krishna Ramanathan: [A Study of Real-World Data Races in Golang](https://arxiv.org/abs/2204.00764), PLDI 2022 (Uber Technologies)
