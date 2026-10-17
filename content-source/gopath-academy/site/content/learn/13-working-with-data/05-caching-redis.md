---
title: "13.5 Caching with Redis"
linkTitle: "13.5 Caching with Redis"
weight: 5
description: "Cache-aside with Redis and go-redis: TTLs, invalidation, request coalescing against cache stampedes (measured), and making the cache optional so a Redis outage never takes the service down."
goVersion: "1.27"
---

## What you will learn

- When a cache helps, and what it costs
- The **cache-aside** pattern with Redis and the go-redis client
- TTLs and invalidation: the two ways cached data stops being stale
- **Cache stampedes**, and coalescing concurrent misses with `singleflight`, measured
- Designing the cache so that a Redis outage degrades performance, not availability

## When to cache

A cache keeps copies of data somewhere faster than its source. It helps when the same data is read **much more often than it changes**, and reading it from the source is slow or expensive: a popular product page, a user's permissions, a rendered fragment, or a short link that will be followed thousands of times.

It costs you complexity and **staleness**: from the moment you cache a value, readers may see an old one. So the first questions are "how stale may this be?" and "is the database actually the bottleneck?" (measure it, Module 10). Databases are fast at reading rows by primary key; a cache in front of a 1 ms query rarely pays for itself.

**Redis** is an in-memory data store used as a shared cache by many instances of a service. `github.com/redis/go-redis/v9` is the official Go client; like `*sql.DB`, a `*redis.Client` is a concurrency-safe connection pool you create once.

## Cache-aside

The most common pattern, also called lazy loading: the application checks the cache, and on a miss loads from the source and fills the cache.

{{< example file="examples/thirdparty/cache/rediscache/cache.go" region="cache" >}}

- **A hit** returns immediately.
- **`redis.Nil`** is go-redis's "key does not exist" error. It is a miss, not a failure.
- **Every entry has a TTL** (`Set(ctx, key, value, ttl)`): Redis deletes it when it expires, so stale data has a maximum age, and unused keys do not fill memory forever.
- **Keys are namespaced** (`prefix`), such as `shortener:link:abc`, so several applications or data types can share one Redis safely.

The test shows the timing:

```text
call 1: 52ms, loads so far 1
call 2: 100µs, loads so far 1
call 3: 100µs, loads so far 1
```

The first call pays for the slow load; the next ones take about a tenth of a millisecond, a round trip to Redis on the same machine. After the TTL, or after `Invalidate`, the next call loads again.

## Invalidation

There are two ways to deal with data that changes:

- **Expiry only.** Accept up to one TTL of staleness. Simple, and often enough: a product description that is five minutes old is fine.
- **Invalidate on write.** When the source changes, delete the cached key (`Invalidate`). Delete rather than update: concurrent writers updating the cache can leave it with the older value. Even then, keep a TTL as a safety net.

The famous difficulty of cache invalidation comes from data that changes often and must never be stale. For that, reconsider caching at all. Data that **never changes**, like a short link, is the easy case: there is nothing to invalidate.

## Cache stampedes

A popular key expires. In the next millisecond, a hundred requests miss at once, and all hundred run the expensive load against the database, which was exactly what the cache was protecting. This is a **cache stampede** (or thundering herd):

{{< example file="examples/thirdparty/cache/rediscache/cache_test.go" region="stampede" run="cd examples/thirdparty && go test -run TestStampede -v ./cache/rediscache" >}}

```text
coalesce=false 100 concurrent misses caused 100 loads
coalesce=true  100 concurrent misses caused 1 loads
```

`golang.org/x/sync/singleflight` fixes it within one process: concurrent calls to `group.Do(key, fn)` with the same key wait for a single execution of `fn` and share its result. Across many instances, each instance does at most one load per key, which is usually good enough. (Distributed locks in Redis can reduce it further, at a price in complexity.)

Note `context.WithoutCancel(ctx)` inside `Do`: the shared load must not be cancelled because the **first** caller's request was cancelled, since other callers are waiting for it too. Exercise 13.6 builds this coalescing by hand.

## The cache is optional

A cache must make the service faster, never less available. If Redis goes down and every request fails, you have added a single point of failure. `Get` therefore treats Redis errors like misses: it logs a warning and loads from the source.

{{< example file="examples/thirdparty/cache/rediscache/cache_test.go" region="down" >}}

Two details make this work in practice: **short timeouts** for Redis (a cache that takes two seconds to fail is worse than no cache), and making sure the source can survive the full load for a while, or protecting it with rate limits (lesson 7.5), because during an outage every request reaches it.

## Common mistakes

### Caching without a TTL

Memory fills up, and stale data lives forever.

### Updating the cache on write

Concurrent writers can leave the old value. Delete instead.

### Treating Redis errors as request failures

The cache becomes a dependency. Fall back to the source.

### Caching errors or "not found" by accident

A brief database error, cached for an hour, becomes an hour of failures. Cache successes; if you cache "not found" (negative caching), use a short TTL.

### Unbounded key space

Keys built from user input (search queries, full URLs) can create millions of entries. Bound them, and set a Redis `maxmemory` with an eviction policy.

## Key takeaways

- Cache data that is read far more often than it changes, after measuring that the source is the bottleneck.
- Cache-aside: check Redis, on `redis.Nil` load from the source and `Set` with a TTL.
- Invalidate by deleting keys; keep TTLs as a safety net.
- Coalesce concurrent misses with `singleflight` to prevent stampedes: measured, 100 loads became 1.
- Redis failures must fall back to the source, with short timeouts.

## Further reading

- [go-redis documentation](https://redis.uptrace.dev/)
- [Package golang.org/x/sync/singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)
- [Redis: key eviction](https://redis.io/docs/latest/develop/reference/eviction/)
