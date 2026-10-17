---
title: "7.7 Graceful shutdown"
linkTitle: "7.7 Graceful shutdown"
weight: 7
description: "Stop a Go service cleanly: catch SIGINT and SIGTERM with signal.NotifyContext, stop accepting work, let in-flight work finish, and respect a deadline that fits inside Kubernetes' grace period."
goVersion: "1.27"
---

## What you will learn

- Why services must shut down gracefully, especially on Kubernetes
- Turning OS signals into a context with `signal.NotifyContext`
- Shutting down an HTTP server with `Shutdown`
- Draining background workers
- Choosing shutdown deadlines

## The idea

Services are stopped all the time: every deploy, every autoscaling event, every node drain. If a process simply dies, in-flight requests fail, half-written data is lost and clients see errors. A **graceful shutdown**:

1. stops accepting new work,
2. lets work already in progress finish,
3. releases resources (flushes logs, closes database connections),
4. and gives up after a deadline, so a stuck request cannot block the shutdown forever.

## Signals to context

Kubernetes, systemd and Docker stop a process by sending **SIGTERM**; pressing Ctrl+C sends **SIGINT**. `signal.NotifyContext` turns them into a cancelled context, which composes with everything from lesson 7.1:

{{< example file="examples/m07/shutdown/main_example.go" region="main" >}}

After the first signal, `stop()` restores the default behavior, so a second Ctrl+C kills the process immediately. That is a good escape hatch when shutdown hangs during development.

## Shutting down an HTTP server

{{< example file="examples/m07/shutdown/shutdown_test.go" region="server" >}}

- `srv.Serve` runs in a goroutine, and its error goes to a buffered channel.
- `srv.Shutdown(ctx)` closes the listeners (no new connections), closes idle connections, and **waits for active requests to finish**, or until `ctx` expires.
- After `Shutdown`, `Serve` returns `http.ErrServerClosed`, which means "stopped on purpose", not a failure.

The demo starts a slow request, triggers shutdown while it is still running, and shows that the request completes:

{{< example file="examples/m07/shutdown/shutdown_test.go" region="demo" run="go test -race -run Example_gracefulShutdown -v ./examples/m07/shutdown" >}}

Note: `Shutdown` does not wait for hijacked connections such as WebSockets. Use `srv.RegisterOnShutdown` to notify them.

## Draining workers

The same idea applies to background workers:

1. Cancel the context the workers watch, or close their jobs channel, so they stop taking new work.
2. `Wait` on their `WaitGroup` (or `errgroup`) with a deadline.
3. Then close shared resources such as database pools, **after** everything using them has stopped.

The order matters: close resources last, in the reverse order of creating them.

## Deadlines and Kubernetes

When Kubernetes deletes a pod, it:

1. removes the pod from Service endpoints (new traffic stops arriving, eventually),
2. runs any `preStop` hook,
3. sends SIGTERM,
4. waits up to `terminationGracePeriodSeconds` (30 seconds by default),
5. sends SIGKILL.

Two consequences:

- Your shutdown deadline must be **shorter** than the grace period, or the process is killed mid-shutdown.
- Endpoint removal is not instant. Load balancers may still send requests for a few seconds after SIGTERM. A common practice is a short `preStop` sleep (a few seconds) or a delay in your handler before calling `Shutdown`, and failing readiness checks as soon as shutdown begins. Lesson 15.2 covers probes and these settings in detail.

## Common mistakes

### Calling `os.Exit` or `log.Fatal` during shutdown

They skip deferred cleanup. Return from `main` instead.

### No deadline on `Shutdown`

A single hung request (or a client that never closes its connection) blocks shutdown until SIGKILL arrives.

### Treating `http.ErrServerClosed` as an error

It is the normal result of `Shutdown`. Check for it with `errors.Is`.

### Closing the database before workers finish

Workers then fail with "database is closed" errors during every deploy. Stop the users of a resource before the resource.

## Key takeaways

- Graceful shutdown means: stop accepting, finish in-flight work, release resources, give up at a deadline.
- `signal.NotifyContext` turns SIGINT and SIGTERM into context cancellation.
- `http.Server.Shutdown(ctx)` waits for active requests; `Serve` then returns `http.ErrServerClosed`.
- Drain workers before closing the resources they use.
- Keep the shutdown deadline below Kubernetes' termination grace period.

## Further reading

- [Package os/signal: NotifyContext](https://pkg.go.dev/os/signal#NotifyContext)
- [net/http: Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown)
- [Kubernetes: Termination of Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination)
