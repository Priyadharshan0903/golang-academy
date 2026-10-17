---
title: "13.6 Exercise: a read-through cache"
linkTitle: "13.6 Exercise: read-through cache"
weight: 6
description: "Build an in-memory read-through cache with expiry, no caching of errors, and request coalescing, without holding a lock during loads; tested with a fake clock and the race detector."
goVersion: "1.27"
---

## The task

Lesson 13.5 used Redis for storage and `singleflight` for coalescing. Here you build both ideas yourself, in memory and with only the standard library, which is the best way to understand what those packages do.

{{< exercise dir="m13/readthrough" >}}
Implement `Cache.Get` so that it follows the five rules in the comment on `Cache`: hits skip the loader, values expire after the TTL, errors are not cached, concurrent misses for one key share a single load, and a slow load never blocks other keys.
{{< /exercise >}}

```bash
cd exercises
go test -race ./m13/readthrough
```

The starter passes two of the four tests by accident (it never caches anything, so it cannot cache an error or block another key). Make all four pass, with `-race`.

## Hints

<details>
<summary>Hint 1: the data</summary>

Two maps behind one `sync.Mutex`: `entries` (value and expiry time) and `inflight` (loads in progress). A load in progress is a small struct with a `done` channel and fields for the result.

</details>

<details>
<summary>Hint 2: the lock</summary>

Rule 5 means you must **not** hold the mutex while calling the loader. Lock, look in both maps, register a new in-flight call if needed, unlock, load, then lock again to store the result and remove the in-flight entry.

</details>

<details>
<summary>Hint 3: waking the waiters</summary>

Closing a channel wakes every goroutine blocked on receiving from it (lesson 6.2). Write the result into the call struct **before** closing `done`; the close makes the writes visible to the waiters (lesson 6.7). Waiters should also give up if their own context ends.

</details>

<details>
<summary>Hint 4: the clock</summary>

Never call `time.Now()` directly; use the `now` function you were given, so the test's fake clock controls expiry.

</details>

The reference solution is in `exercises/m13/readthrough/solution.go`. Compare it with `singleflight`'s source: the core idea is the same.
