---
title: "7.5 Rate limiting and backpressure"
linkTitle: "7.5 Rate limiting"
weight: 5
description: "Limit how much work happens at once and how fast: channel semaphores, ticker spacing, the token bucket algorithm, golang.org/x/time/rate, and bounded queues that push back on producers."
goVersion: "1.27"
---

## What you will learn

- The difference between limiting **concurrency** and limiting **rate**
- A buffered channel as a counting semaphore
- Spacing work out with a ticker
- How the **token bucket** algorithm works, and the `x/time/rate` package
- Backpressure: making producers slow down or fail fast instead of piling up work

## The idea

Two different limits protect systems:

- **Concurrency limit**: at most N things **at the same time** (database connections, open files, in-flight requests).
- **Rate limit**: at most N things **per unit of time** (an API that allows 100 requests per second).

And when work arrives faster than you can handle it, you need **backpressure**: a way for the slow part to push back on the fast part, instead of queuing without limit until memory runs out.

## Limiting concurrency: a channel semaphore

{{< example file="examples/m07/ratelimiting/ratelimiting_test.go" region="semaphore" run="go test -race -run Example_semaphore -v ./examples/m07/ratelimiting" >}}

Sending into a buffered channel of capacity N "acquires" a slot, receiving "releases" it. When all N slots are taken, the next send blocks. It is the same mechanism inside `errgroup.SetLimit`. For weighted limits (some tasks take several slots), use `golang.org/x/sync/semaphore`.

## Limiting rate: a ticker

{{< example file="examples/m07/ratelimiting/ratelimiting_test.go" region="ticker" run="go test -run Example_tickerRateLimit -v ./examples/m07/ratelimiting" >}}

Waiting for a tick before each request gives a steady rate. It is simple, but it allows **no bursts**: even after an idle minute, the next request still waits for a tick.

## The token bucket

Most real rate limiters use a **token bucket**:

- The bucket holds up to `burst` tokens and starts full.
- Tokens are added at `rate` per second, never beyond `burst`.
- Each event takes one token. With no token available, the event is rejected or waits.

This allows short bursts (up to `burst` events at once after an idle period) while keeping the long-term average at `rate`. The standard implementation is `golang.org/x/time/rate`:

{{< example file="examples/thirdparty/ratedemo/ratedemo_test.go" region="rate" run="cd examples/thirdparty && go mod tidy && go test ./ratedemo" >}}

- `Allow()` answers "may I, right now?" without blocking. Good for rejecting requests with HTTP 429.
- `Wait(ctx)` blocks until a token is available, or returns an error if `ctx` ends first. Good for outbound calls you want to slow down rather than drop.
- `Reserve()` tells you how long you would have to wait, for schedulers.

In lesson 7.8 you will look at a token bucket written from scratch with only the standard library, including a per-key limiter for "N requests per second **per client**".

## Backpressure

{{< example file="examples/m07/ratelimiting/ratelimiting_test.go" region="backpressure" run="go test -run Example_backpressure -v ./examples/m07/ratelimiting" >}}

A bounded queue gives three choices when it is full:

| Strategy | Code | When |
|---|---|---|
| **Block** the producer | plain send | Batch jobs where waiting is fine |
| **Fail fast** | `select` with `default` | Request handling: return 429 or 503 and let the client retry |
| **Wait with a deadline** | `select` with `ctx.Done()` | Callers that can wait a little but not forever |

What you should **not** do is accept unlimited work. An unbounded queue turns a short overload into an outage: memory grows, latency grows, and by the time work is processed, the clients have long given up.

## Common mistakes

### Confusing concurrency limits with rate limits

A semaphore of 10 can still make thousands of requests per second if each is fast. If the other side limits rate, limit rate.

### One limiter per request

Creating a new `rate.Limiter` inside each handler call means every request gets a full bucket. The limiter must be shared (per service, or per key).

### Per-key limiters that never forget keys

A map of limiters per IP address grows forever. Prune idle keys (the `Keyed` limiter in lesson 7.8 does this).

### Using `time.Sleep` for rate limiting

`time.Sleep(10 * time.Millisecond)` between calls ignores cancellation and drifts. Use a ticker or a token bucket with `Wait(ctx)`.

## Key takeaways

- Limit concurrency with a semaphore; limit rate with a ticker or token bucket.
- A buffered channel of capacity N is a counting semaphore.
- Token buckets allow bursts up to `burst` while averaging `rate`; `x/time/rate` is the standard implementation.
- Bound every queue, and decide whether producers block, fail fast or wait with a deadline.

## Further reading

- [Package golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)
- [Token bucket](https://en.wikipedia.org/wiki/Token_bucket) on Wikipedia
