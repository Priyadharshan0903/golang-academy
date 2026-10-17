---
title: "7.8 Projects: token bucket rate limiter and concurrent crawler"
linkTitle: "7.8 Projects"
weight: 8
description: "Two complete concurrent programs: a token bucket rate limiter written from scratch, and a polite web crawler with a coordinator goroutine, a worker pool and cancellation."
goVersion: "1.27"
---

These two projects put Modules 6 and 7 together. The crawler uses the rate limiter, so read them in order.

## Project 1: a token bucket rate limiter

Code: [`projects/ratelimit`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/ratelimit)

Lesson 7.5 introduced the token bucket and used `golang.org/x/time/rate`. This project implements the same algorithm with only the standard library, so you can see every detail.

{{< example file="projects/ratelimit/ratelimit.go" >}}

Design points:

- **Lazy refill.** There is no background goroutine adding tokens. Each call computes how many tokens accumulated since the last call (`elapsed * rate`, capped at `burst`). An idle bucket costs nothing.
- **One mutex** guards all fields. Every method locks, refills, and acts.
- **`WaitN` reserves before sleeping.** It subtracts the tokens immediately (the balance may go negative) and sleeps until they would be earned. That makes concurrent waiters fair: each is scheduled after the ones before it. If the context ends first, the reservation is returned.
- **Fail fast on impossible deadlines.** If the required wait is longer than the time left on the context, `WaitN` returns immediately instead of sleeping and failing later.
- **Injected clock.** The `now` field lets tests advance time manually and check exact token counts.

### Per-key limiting

{{< example file="projects/ratelimit/keyed.go" >}}

`Keyed` gives every key (a client IP, an API key) its own bucket, created on first use. `Prune` deletes buckets that have been idle, so memory does not grow with every client ever seen. In Module 12 this becomes HTTP middleware that returns 429 Too Many Requests.

### Tests

{{< example file="projects/ratelimit/ratelimit_test.go" >}}

With a fake clock, refill math can be checked exactly, and 500 concurrent goroutines get **exactly** 100 tokens. On Go 1.25 and later, [`synctest_test.go`](https://github.com/Priyadharshan0903/gopath-academy/blob/main/projects/ratelimit/synctest_test.go) also checks `Wait`'s real timer behavior on a virtual clock. The benchmarks are used again in Module 10.

## Project 2: a concurrent web crawler

Code: [`projects/crawler`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/crawler)

The crawler starts at a URL, follows links on the same host up to a depth limit, and reports every page's status.

### Architecture

```text
              tasks (unbuffered)
coordinator  ───────────────────▶  worker 1 ─┐
  owns:      ───────────────────▶  worker 2 ─┤  fetch (rate limited)
  - queue    ───────────────────▶  worker N ─┘
  - seen set ◀───────────────────  outcomes (unbuffered)
  - results
```

- **One coordinator goroutine owns all state**: the queue, the set of seen URLs and the results. Nothing is shared, so there are no mutexes at all; channels are the only synchronization.
- **A fixed pool of workers** only fetches pages and reports outcomes.
- **The coordinator's loop** selects between "send the next task" and "receive an outcome". The send case uses the **nil channel trick** from lesson 6.3: when the queue is empty, the send channel is `nil`, which disables that case.
- **Termination**: the crawl is done when the queue is empty **and** nothing is in flight. This is the classic subtle part of crawlers: an empty queue alone is not enough, because a running fetch may discover new links.

{{< example file="projects/crawler/internal/crawl/crawl.go" >}}

### Fetching

{{< example file="projects/crawler/internal/crawl/fetch.go" >}}

- The `Fetcher` interface keeps the crawler testable.
- Every request carries the context, so cancelling the crawl aborts in-flight HTTP requests too.
- The body is read through `io.LimitReader`, so a huge page cannot exhaust memory.
- Links are found with a regular expression. That is **not** a correct HTML parser, and the code says so; a production crawler should use `golang.org/x/net/html`. It keeps the project dependency-free.

### Testing without the internet

{{< example file="projects/crawler/internal/crawl/crawl_test.go" >}}

`httptest.NewServer` serves a small fake site on localhost with a cycle (`/a` links back to `/`), a fragment link to an already seen page, a broken link, an external link and a slow page. The tests check that every page is fetched **exactly once** (the server counts hits), that depth and page limits work, that a timeout cancels the crawl, and that no goroutines leak.

### Running it

```bash
make serve                                                  # in one terminal: the course site
go run ./projects/crawler/cmd/crawler -url http://localhost:1313/ -depth 2 -rate 20
```

Only crawl sites you own or have permission to crawl, and keep the rate low. Being polite is part of the design.

## Stretch goals

1. Respect `robots.txt` before crawling a site.
2. Add per-host rate limiting with `ratelimit.Keyed` so a crawler following several hosts is polite to each.
3. Replace the regular expression with `golang.org/x/net/html` in the `examples/thirdparty` module style (a separate module with the dependency).
4. Stream results to the caller through an `iter.Seq[Result]` instead of returning a slice at the end.

## Part 2, first half: recap

You can now:

- Start, wait for and stop goroutines without leaking them
- Use channels, `select`, mutexes and atomics, and know when to pick each
- Find data races with `-race` and explain the memory model's guarantees
- Propagate cancellation with `context`
- Build worker pools, pipelines and error groups
- Rate limit and apply backpressure
- Shut a service down gracefully

Next is **[Module 8: Standard library deep dives](/learn/08-standard-library/)**.
