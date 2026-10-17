---
title: "15.6 Exercise: drain gracefully"
linkTitle: "15.6 Exercise: drain gracefully"
weight: 6
description: "Implement the shutdown sequence that took a rolling update from 59 failed requests to zero: flip readiness, keep serving for a delay, then shut down with a timeout."
goVersion: "1.27"
---

## The task

Lesson 15.2 measured what happens when a service stops the moment it gets SIGTERM: 59 failed requests in one rolling update. Here you write the function that prevents it.

{{< exercise dir="m15/drain" >}}
Implement `Serve` in `starter.go`: serve until the context is cancelled, then mark the service not ready, keep serving for `opts.Delay`, and shut down with `srv.Shutdown`, giving in-flight requests up to `opts.Timeout`.
{{< /exercise >}}

```bash
cd exercises
go test -race ./m15/drain
```

The tests run a real HTTP server and check that:

- readiness turns false **immediately** on cancel, while requests are **still served** during the delay;
- a request in flight when shutdown starts **completes**;
- a request that hangs does not hang `Serve`: it returns `context.DeadlineExceeded` after the timeout.

## Hints

<details>
<summary>Hint 1: two things to wait for</summary>

Run `srv.Serve(ln)` in a goroutine that sends its result on a buffered channel. Then `select` on that channel (the server failed on its own) and on `ctx.Done()` (time to shut down).

</details>

<details>
<summary>Hint 2: the delay</summary>

Use a `select` on a timer and the serve channel, not `time.Sleep`, so a server that dies during the delay is reported at once.

</details>

<details>
<summary>Hint 3: the return value</summary>

After `Shutdown`, `Serve` returns `http.ErrServerClosed`: that is success, not an error. Receive it from the channel so the goroutine does not leak, and return what `Shutdown` returned.

</details>

The reference solution is in `exercises/m15/drain/solution.go`. Compare it with the shortener's `run` in `projects/shortener/cmd/shortener/main.go`, which does the same for two servers at once.
