---
title: "7.4 errgroup and bounded concurrency"
linkTitle: "7.4 errgroup"
weight: 4
description: "Run a group of goroutines where the first error cancels the rest, limit how many run at once, and see how golang.org/x/sync/errgroup works by building a small version of it."
goVersion: "1.27"
---

## What you will learn

- The "first error cancels everything" pattern
- How `golang.org/x/sync/errgroup` implements it, by building a version yourself
- How `SetLimit` bounds concurrency with a semaphore
- How to use the real `errgroup` package
- Its limits, and when to choose something else

## The idea

A very common shape of concurrent work: start N tasks, wait for all of them, and if **any** fails, stop the others and report the error. Writing it by hand each time means a `WaitGroup`, a `sync.Once` for the first error, a cancellable context and a semaphore. The `errgroup` package, maintained by the Go team in `golang.org/x/sync`, wraps all of that.

## Building it from the standard library

Here is a compact version with the same API, so you can see exactly what happens inside:

{{< example file="examples/m07/boundedgroup/group.go" region="group" >}}

- `WithContext` derives a cancellable context. The group cancels it, with the first error as the cause, as soon as any function fails.
- `sync.Once` keeps only the **first** error.
- `SetLimit` creates a buffered channel used as a **semaphore**: `Go` must put a token in before starting a goroutine, so when `n` are running, the next `Go` call blocks.
- `Wait` waits for everything, then cancels the context to release resources.

Using it:

{{< example file="examples/m07/boundedgroup/group_test.go" region="use" run="go test -race -run Example_firstErrorCancelsOthers -v ./examples/m07/boundedgroup" >}}

One task fails, the other four see `ctx.Done()` and return early, and `Wait` returns the original failure.

{{< example file="examples/m07/boundedgroup/group_test.go" region="limit" run="go test -race -run Example_limit -v ./examples/m07/boundedgroup" >}}

## Using the real package

In real code, use `golang.org/x/sync/errgroup` rather than your own copy. It lives in this course's separate [`examples/thirdparty`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/examples/thirdparty) module, so the main course module stays dependency-free:

{{< example file="examples/thirdparty/errgroupdemo/errgroupdemo_test.go" region="errgroup" run="cd examples/thirdparty && go mod tidy && go test ./errgroupdemo" >}}

The `x/sync/errgroup` API also has `TryGo`, which starts a goroutine only if the limit has room and reports whether it did, and the zero `errgroup.Group` (without `WithContext`) when you want first-error collection without cancellation.

## Limits of errgroup

- **Only the first error.** If you need every error, collect them yourself and join them (lesson 5.4).
- **Results are up to you.** `Go` takes `func() error`. Write results into an indexed slice or guard a shared structure with a mutex, as `fetchAll` does.
- **Panics are not converted to errors.** A panic in a group goroutine crashes the program like any other goroutine panic (lesson 5.5). Recover inside the function if the code might panic.
- **Cancellation is cooperative.** The context is cancelled, but each function must actually check it.

## Common mistakes

### Ignoring the derived context

Using the outer `ctx` instead of the one returned by `WithContext` means failures cannot cancel the other tasks.

### Shared results without synchronization

Appending to a shared slice from several `g.Go` functions without a mutex is a data race. Use indexes or a lock.

### No limit on huge inputs

`g.Go` for a million items without `SetLimit` starts a million goroutines. Set a limit.

## Key takeaways

- `errgroup` runs functions concurrently, waits for all, returns the first error and cancels the rest.
- It is built from a `WaitGroup`, a `sync.Once`, a cancellable context and an optional semaphore.
- `SetLimit` bounds concurrency; `TryGo` starts work only if there is room.
- Collect results with indexes or a mutex, recover panics yourself, and pass the derived context to every task.

## Further reading

- [Package golang.org/x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [Package golang.org/x/sync/semaphore](https://pkg.go.dev/golang.org/x/sync/semaphore) for weighted semaphores
