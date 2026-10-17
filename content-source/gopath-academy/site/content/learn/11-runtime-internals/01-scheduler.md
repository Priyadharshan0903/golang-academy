---
title: "11.1 How the scheduler works: G, M and P"
linkTitle: "11.1 The scheduler"
weight: 1
description: "How the Go runtime runs many goroutines on few OS threads: Gs, Ms and Ps, run queues and work stealing, blocking system calls, preemption, and how GOMAXPROCS is chosen inside containers."
goVersion: "1.27"
---

## What you will learn

- The three scheduler objects, **G**, **M** and **P**, and how they fit together
- Run queues, work stealing, and what happens when a goroutine blocks
- How preemption keeps a busy goroutine from starving others
- What `GOMAXPROCS` controls, and how Go 1.25 picks it inside containers
- How to observe the scheduler with `runtime/metrics` and `go test -cpu`

## The idea

You can start a million goroutines, but the machine has only a few CPU cores and the operating system only so many threads. The Go **scheduler**, part of the runtime linked into every Go binary, multiplexes goroutines onto threads. It uses three kinds of objects:

| Object | What it is |
|---|---|
| **G** (goroutine) | A goroutine: its stack, its saved registers, its status (running, runnable, waiting). Cheap: a few KiB. |
| **M** (machine) | An operating system thread. Ms execute code. |
| **P** (processor) | A scheduling context: the right to run Go code, plus a local queue of runnable Gs. There are exactly `GOMAXPROCS` Ps. |

An M must hold a P to run Go code. So `GOMAXPROCS` limits how many goroutines **execute Go code at the same moment**. It does not limit how many goroutines exist, and it does not limit threads: threads blocked in system calls do not hold a P.

```text
        P0                    P1                  global run queue
   ┌──────────┐          ┌──────────┐            ┌──────────────┐
   │ runq:    │          │ runq:    │            │ G G G ...    │
   │ G G G G  │          │ G        │            └──────────────┘
   └────┬─────┘          └────┬─────┘
        │ held by             │ held by
        M0 ── running G7      M1 ── running G2        M2: blocked in a syscall, no P
```

## Run queues and work stealing

Each P has a **local run queue** of up to 256 runnable goroutines, and there is one **global run queue** shared by all. When a goroutine is created with `go`, it usually goes onto the current P's local queue (in a special "run next" slot, so a newly created goroutine runs soon, which is good for producer and consumer pairs).

When a P runs out of work, it looks, roughly in this order: its own queue, the global queue, the network poller (goroutines whose I/O is ready), and then it **steals half** of another P's queue. Work stealing keeps all cores busy without a central lock.

## When goroutines block

- **Channels, mutexes, `time.Sleep`, `select`**: the goroutine is parked (status "waiting") and the M simply picks another G from the P. No thread is blocked. This is why blocking a goroutine is cheap.
- **Network I/O**: also parked. The runtime's **network poller** (epoll on Linux, kqueue on BSD and macOS, IOCP on Windows) watches the sockets and makes goroutines runnable when data arrives. That is how one Go process serves tens of thousands of connections with a handful of threads, while your code reads like simple blocking code.
- **Blocking system calls** (file I/O on most systems, cgo calls): the thread really blocks. The runtime detaches the P from that M so another M can use it to run goroutines, starting a new thread if needed. When the system call returns, the goroutine needs a P again.

A background thread called **sysmon** watches for long system calls and long-running goroutines, and handles both.

## Preemption

What if a goroutine never blocks? This example pins the program to one P and starts a goroutine with a tight loop that makes no function calls:

{{< example file="examples/m11/sched/sched.go" region="preempt" run="go test -v -run TestPreemption ./examples/m11/sched" >}}

On Go 1.27 it prints something like `main goroutine ran again after 11.860807ms`. The main goroutine woke up from its 10 ms sleep, but the only P was busy spinning. sysmon noticed that the spinner had run for longer than its time slice (10 ms) and **preempted** it, so main could run and stop it.

Before Go 1.14 this program hung forever: goroutines could only be preempted at function calls, and this loop has none. Since Go 1.14, the runtime also uses **asynchronous preemption**: it sends the thread a signal (`SIGURG` on Unix) and stops the goroutine wherever it is. You can see the old behavior by turning that off:

```bash
go test -c -o sched.test ./examples/m11/sched
GODEBUG=asyncpreemptoff=1 timeout 10 ./sched.test -test.run Preemption   # hangs until timeout
```

You will never need `asyncpreemptoff` in real code; it exists for debugging.

## GOMAXPROCS

`GOMAXPROCS` is the number of Ps. You can set it with the `GOMAXPROCS` environment variable or `runtime.GOMAXPROCS(n)`, which returns the previous value (`runtime.GOMAXPROCS(0)` just reads it).

The default used to be simply the number of CPU cores the machine has (`runtime.NumCPU`). In a container that is often wrong: a Kubernetes pod with a CPU limit of 2 on a 64-core node got 64 Ps, so the Go program ran up to 64 threads at once, used up its CPU quota early in each scheduling period and was **throttled** by the kernel, causing latency spikes. For years, the fix was the `go.uber.org/automaxprocs` package.

Since **Go 1.25** the runtime handles this itself:

- On Linux, the default `GOMAXPROCS` respects the **cgroup CPU bandwidth limit**, which in Kubernetes is the pod's **CPU limit**. CPU **requests** are not considered. A limit of 2.5 CPUs gives `GOMAXPROCS=3` (rounded up, and at least 2 if the machine has 2 or more CPUs).
- On all systems, the runtime **re-checks periodically** and updates `GOMAXPROCS` if the limit or the number of available CPUs changes.
- Both behaviors are off if you set `GOMAXPROCS` yourself, and can be disabled with the `GODEBUG` settings `containermaxprocs=0` and `updatemaxprocs=0`. `runtime.SetDefaultGOMAXPROCS()` returns to the default.

These behaviors depend on the `go` line in your `go.mod` (GODEBUG defaults follow it), so a module that still says `go 1.24` keeps the old default even when built with a newer toolchain. Module 15 comes back to this when we deploy to Kubernetes.

## Seeing parallelism

`GOMAXPROCS` decides how much of your concurrency becomes parallelism. This function hashes 64 blocks of 64 KiB with 8 goroutines:

{{< example file="examples/m11/sched/sched.go" region="parallel" >}}

{{< example file="examples/m11/sched/sched_test.go" region="bench" >}}

The `-cpu 1,2` flag runs the benchmark once per `GOMAXPROCS` value. On the 2-core test VM (`benchstat` output, 6 runs each):

```text
          │   sec/op     │
HashAll      2.903m ± 47%     (GOMAXPROCS=1)
HashAll-2    1.554m ±  2%     (GOMAXPROCS=2)
```

Same code, same 8 goroutines: twice the Ps, about twice the throughput. Asking for `-cpu 4` on a 2-core machine would not help further: there are only 2 cores to run on. The `-2` suffix that appears on every benchmark name in this course is exactly this value.

## Observing the scheduler

`runtime/metrics` exposes scheduler data without any profiling:

{{< example file="examples/m11/sched/sched.go" region="metrics" >}}

Useful metrics include `/sched/gomaxprocs:threads`, `/sched/goroutines:goroutines`, and `/sched/latencies:seconds`, a histogram of how long runnable goroutines waited for a P. High scheduling latency means more work is runnable than there are Ps to run it. Go 1.26 added counts of goroutines by state under `/sched/goroutines`, `/sched/threads:threads`, and `/sched/goroutines-created:goroutines`. Run `go doc runtime/metrics` for the full list your version supports. For a timeline of scheduling decisions, use the execution tracer (lesson 10.3).

## Common mistakes

### Setting GOMAXPROCS higher "for more concurrency"

Concurrency comes from goroutines. More Ps than cores only adds contention.

### Assuming goroutines run in creation order

The scheduler gives no ordering guarantees. Synchronize explicitly (Modules 6 and 7).

### Ignoring CPU limits on older Go versions

On Go 1.24 and earlier, or with a `go.mod` that says so, a containerized service may run far more Ps than its CPU limit. Upgrade, or set `GOMAXPROCS` explicitly.

## Key takeaways

- Gs are goroutines, Ms are threads, Ps are the right to run Go code; there are `GOMAXPROCS` Ps.
- Each P has a local run queue; idle Ps steal work. Blocking on channels, locks and network I/O parks only the goroutine, not a thread.
- Since Go 1.14, long-running goroutines are preempted asynchronously, even in tight loops.
- Since Go 1.25, the default `GOMAXPROCS` respects the container's CPU limit and updates when it changes.
- Measure scaling with `go test -cpu`, and watch the scheduler with `runtime/metrics` and the tracer.

## Further reading

- [Go 1.25 release notes: container-aware GOMAXPROCS](https://go.dev/doc/go1.25#container-aware-gomaxprocs)
- [Package runtime/metrics](https://pkg.go.dev/runtime/metrics)
- [Scalable Go Scheduler design document](https://golang.org/s/go11sched) by Dmitry Vyukov, the original design of Ps
