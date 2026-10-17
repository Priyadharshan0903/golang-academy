---
title: "6.1 Goroutines and the cost model"
linkTitle: "6.1 Goroutines"
weight: 1
description: "Start concurrent work with the go statement, wait for it correctly, understand why goroutines are cheap, and how GOMAXPROCS controls parallelism."
goVersion: "1.27"
---

## What you will learn

- How to start a goroutine, and why you must always plan how it ends
- How to wait for goroutines with `sync.WaitGroup`
- Why goroutines are cheap compared to OS threads
- The difference between concurrency and parallelism, and what `GOMAXPROCS` does
- What happens to running goroutines when `main` returns

## The idea

A **goroutine** is a function running concurrently with the rest of the program. You start one by putting `go` in front of a function call:

```go
go sendEmail(user) // returns immediately; sendEmail runs concurrently
```

That is the whole syntax. The hard part is everything around it: knowing when the goroutine finishes, getting its results back, and making sure it does not outlive what it was working for. A good rule to adopt from day one: **never start a goroutine without knowing how it will stop.**

## Starting and waiting

{{< example file="examples/m06/goroutines/goroutines_test.go" region="start" run="go test -run Example_startGoroutines -v ./examples/m06/goroutines" >}}

- `wg.Add(1)` before each `go` statement, `defer wg.Done()` inside the goroutine, `wg.Wait()` to block until all are done.
- Each goroutine writes to **its own index** of `results`, so they never touch the same memory. That is why no lock is needed here. (Since Go 1.22, `i` is a fresh variable per iteration, so each closure sees its own value.)
- Goroutines run in **no particular order**. Collecting results by index is one way to get a deterministic result.

## Goroutines are cheap

{{< example file="examples/m06/goroutines/goroutines_test.go" region="cheap" run="go test -run Example_cheap -v ./examples/m06/goroutines" >}}

Ten thousand goroutines start in milliseconds. Why so cheap?

| | OS thread | Goroutine |
|---|---|---|
| Initial stack | typically 1 to 8 MB, fixed | a few KB, grows and shrinks as needed |
| Created and scheduled by | the operating system kernel | the Go runtime, in user space |
| Switching cost | a kernel context switch | a much cheaper switch inside the runtime |
| Practical count | thousands | hundreds of thousands, even millions |

The Go runtime multiplexes many goroutines onto a small number of OS threads (the **M:N scheduler**, covered in lesson 11.1). When a goroutine blocks on a channel, a lock, a sleep or network I/O, the runtime parks it and runs another one on the same thread.

Cheap does not mean free. Every goroutine holds its stack and whatever it references, and a goroutine that never ends is a **memory leak** (lesson 7.6).

## Concurrency vs parallelism

- **Concurrency** is structuring a program as independently executing tasks.
- **Parallelism** is actually running several of them at the same instant, on several CPU cores.

Go gives you concurrency through goroutines; the runtime provides parallelism by running goroutines on up to **`GOMAXPROCS`** threads at once. By default `GOMAXPROCS` is the number of CPUs available.

Since Go 1.25, on Linux the default also respects the **container CPU limit** (the cgroup CPU quota). A Go service in a Kubernetes pod limited to 2 CPUs uses `GOMAXPROCS=2` even on a 64-core node, and the runtime updates it if the limit changes. Before 1.25, people used a library such as `automaxprocs` for this; with current Go you usually do not need it.

## When `main` returns

When `main` returns (or `os.Exit` is called), the program ends **immediately**. Running goroutines are not waited for, and their deferred functions do not run.

{{< example file="examples/m06/goroutines/goroutines_test.go" region="mainexit" run="go test -run Example_notWaiting -v ./examples/m06/goroutines" >}}

The work in the goroutine is simply lost. In a real program, any goroutine doing something important (writing a file, sending a request, flushing logs) must be waited for before `main` returns. Lesson 7.7 covers graceful shutdown.

## Common mistakes

### Calling `wg.Add` inside the goroutine

```go
go func() {
	wg.Add(1) // too late: Wait may already have returned
	defer wg.Done()
}()
```

Always `Add` before the `go` statement. `go vet` has a `waitgroup` analyzer (Go 1.25+) that reports this.

### Fire and forget

`go doSomething()` with no way to know when it ends, whether it failed, or how to stop it. Every goroutine needs an owner who waits for it and a way to signal it to stop (`context`, lesson 7.1).

### Using `time.Sleep` to "wait for goroutines"

Sleeping makes tests slow and flaky. Use a `WaitGroup`, a channel, or `testing/synctest` (lesson 6.3).

### Unbounded goroutines per request

Starting one goroutine per item for a million items works on your laptop and falls over in production. Bound concurrency with a worker pool (lesson 7.2).

## Key takeaways

- `go f()` starts `f` concurrently and returns at once.
- Wait for goroutines with `sync.WaitGroup`: `Add` before `go`, `Done` in a `defer`, then `Wait`.
- Goroutines start small and are scheduled by the runtime, so you can run very many of them.
- `GOMAXPROCS` limits parallelism; since Go 1.25 it respects container CPU limits on Linux.
- When `main` returns, everything stops. Plan how every goroutine ends.

## Further reading

- [Effective Go: Goroutines](https://go.dev/doc/effective_go#goroutines)
- [Concurrency is not parallelism](https://go.dev/blog/waza-talk), Rob Pike's talk
- [Container-aware GOMAXPROCS](https://go.dev/blog/container-aware-gomaxprocs) on the Go blog
- [Go 1.25 release notes](https://go.dev/doc/go1.25)
