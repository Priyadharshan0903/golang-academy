---
title: "6.6 The race detector"
linkTitle: "6.6 Race detector"
weight: 6
description: "What a data race is, how to find races with go test -race, how to read a race report, and the standard ways to fix them."
goVersion: "1.27"
---

## What you will learn

- What exactly a data race is, and why it is worse than "a wrong number sometimes"
- How to run the race detector, and where to run it
- How to read a race report
- The three standard fixes
- The race detector's limits

## The idea

A **data race** happens when two goroutines access the same memory location concurrently, at least one access is a write, and nothing synchronizes them.

Races are not just "occasionally wrong results". The Go memory model (lesson 6.7) gives **no guarantees** for racy programs: you can see torn values, stale values that never update, and corrupted data structures. Some races on maps crash the program. Races are also notoriously hard to reproduce, because they depend on timing.

Go ships a **race detector** built on Google's ThreadSanitizer. It instruments memory accesses at compile time and reports races as they happen at run time.

## Running it

Add `-race` to `go test`, `go run` or `go build`:

```bash
go test -race ./...
go run -race ./cmd/server
go build -race -o server-race ./cmd/server
```

Here is a deliberate race. It lives behind a build tag, so it never runs in normal builds:

{{< example file="examples/m06/racedetector/demo/racy_test.go" region="racy" run="go test -race -tags racedemo ./examples/m06/racedetector/demo" >}}

## Reading a race report

```text
==================
WARNING: DATA RACE
Read at 0x00c000012118 by goroutine 9:
  .../demo.TestRacyCounter.func1()
      .../racy_test.go:23 +0x84

Previous write at 0x00c000012118 by goroutine 8:
  .../demo.TestRacyCounter.func1()
      .../racy_test.go:23 +0x96

Goroutine 9 (running) created at:
  .../demo.TestRacyCounter()
      .../racy_test.go:20 +0x...
==================
--- FAIL: TestRacyCounter
    testing.go: race detected during execution of test
```

Three parts to look at:

1. **The two conflicting accesses**: what kind (read or write), which goroutine, and the stack trace down to the file and line. Here, both are line 23, `n++`.
2. **Where each goroutine was created**: tells you which `go` statements started the conflicting goroutines.
3. **The test fails.** With `-race`, a detected race fails the test even if every assertion passed.

## The three standard fixes

{{< example file="examples/m06/racedetector/fixed_test.go" region="fixed" run="go test -race -v ./examples/m06/racedetector" >}}

1. **Lock it**: guard every access with a `sync.Mutex`.
2. **Make it atomic**: for a single value, use `sync/atomic`.
3. **Do not share it**: give the data to one goroutine and send it messages over a channel.

A fourth fix, often the best: **restructure so goroutines do not share the variable at all.** Give each goroutine its own slot (as `Sum` did in the lesson 6.4 exercise) and combine results after `Wait`.

## Where to run it

- **Always in CI**: `go test -race ./...`. This course's CI does so.
- **Locally** while developing concurrent code.
- **In a canary or staging environment** with a `-race` build, if you can afford the overhead, to catch races that tests do not exercise.

The race detector typically slows programs down 2 to 20 times and uses 5 to 10 times more memory, so it is not meant for normal production builds.

## Limits

- It only finds races in code that **actually runs** during the test. A race in an untested branch stays hidden. Good concurrent tests exercise the concurrent paths.
- It finds data races, **not** logical race conditions (two correctly locked operations in the wrong order) and not deadlocks.
- It requires cgo on most platforms and is supported on common 64-bit systems (see the docs for the full list).

## Common mistakes

### Running tests without `-race`

Concurrent code that "passes" without `-race` may simply not have lost a race yet. Make `-race` the default for `go test` on concurrent packages.

### "Benign" races

"It is only a stats counter, a wrong value is fine." The memory model gives no such guarantee, and the compiler may optimize racy code in surprising ways. Fix it; atomics are cheap.

### Silencing the detector

There is no supported way to mark a race as acceptable. If it reports something, there is a real unsynchronized access.

## Key takeaways

- A data race is a concurrent unsynchronized access where at least one side writes; racy programs have no guarantees.
- Run `go test -race ./...` in CI and while developing.
- Race reports show both conflicting accesses and where their goroutines started.
- Fix with a mutex, an atomic, a channel, or by not sharing the data.
- The detector only sees code paths that run.

## Further reading

- [Data Race Detector](https://go.dev/doc/articles/race_detector) on go.dev
- [Introducing the Go Race Detector](https://go.dev/blog/race-detector) on the Go blog
