---
title: "5.7 Project milestone: URL shortener core"
linkTitle: "5.7 Project: shortener core"
weight: 7
description: "Milestone 1 of the URL shortener: a domain package with sentinel errors, validation, a Store interface, an in-memory store and a fully tested Service."
goVersion: "1.27"
---

## The project

Over the rest of the course you will build a URL shortener in milestones, each one using the module you just finished:

| Milestone | After module | Adds |
|---|---|---|
| **1. Core** (this lesson) | 5. Errors | Domain types, errors, validation, `Service`, in-memory store |
| 2. HTTP API | 12. HTTP services | JSON API, redirects, middleware |
| 3. Postgres | 13. Working with data | A database-backed `Store`, migrations |
| 4. Deploy | 15. Cloud native | Container image and Kubernetes manifests |

The code lives in [`projects/shortener`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/shortener).

## The design

The core is a package called `link` with **no** knowledge of HTTP or databases:

- A `Link` struct.
- **Sentinel errors** callers can check: `ErrNotFound`, `ErrCodeTaken`, `ErrInvalidURL`, `ErrInvalidCode`.
- A small `Store` interface (two methods in this milestone; milestone 4 adds a third, `Delete`), defined in the package that uses it (lesson 4.8).
- A `Service` holding the rules: validate input, generate codes, retry on collisions.

{{< example file="projects/shortener/internal/link/link.go" >}}

### Things to notice

- **Wrapping with context, keeping causes inspectable.** Errors look like `create "my-link": code already in use`, and `errors.Is(err, ErrCodeTaken)` still works through the wrapper.
- **Validation errors carry detail.** `ErrInvalidURL` is wrapped with `%w` alongside the specific reason, so the HTTP layer can return a helpful message and a 400 status.
- **Retry only the retryable error.** A random code collision (`ErrCodeTaken`) is retried; any other store error is returned immediately. After five collisions, the final error says so and still wraps `ErrCodeTaken`.
- **Dependencies are fields.** `now` and `rand` are function fields, set to real implementations by `NewService`. Tests replace them to get fixed timestamps and to force collisions, without any mocking library.
- **Unbiased random codes.** `randomCode` uses `crypto/rand` (unguessable) and rejects bytes that would cause modulo bias. The alphabet leaves out characters that look alike.
- **The context parameter** is part of the `Store` interface from the start, because the database store in milestone 3 will need it. Module 7 explains it.

## The in-memory store

{{< example file="projects/shortener/internal/link/memstore.go" >}}

The `sync.RWMutex` makes the store safe for concurrent use, which matters once the HTTP server handles requests in parallel. Module 6 explains mutexes; for now, note that every method locks before touching the map.

## The tests

{{< example file="projects/shortener/internal/link/link_test.go" >}}

Every error path is tested with `errors.Is`, including the collision retry logic, which would be very hard to trigger with real randomness.

```bash
go test -race ./projects/shortener/...
```

## Stretch goals

1. Add an `ErrReservedCode` for codes like `api`, `admin` and `health`, which the HTTP layer will need for its own routes. Return it from `validateCode`.
2. Add an optional expiry time to `Link`, and make `Resolve` return a new `ErrExpired` for expired links.
3. Write a `CountingStore` decorator (like the `kv` exercise in lesson 4.2) that counts resolves per code.

## Module 5 recap

You can now:

- Create, return and handle errors idiomatically
- Wrap errors with context and inspect them with `errors.Is`, `errors.As` and `errors.AsType`
- Choose between sentinels, error types and behavior checks
- Report several errors at once with `errors.Join`
- Use `panic` and `recover` only where they belong
- Structure error handling across the layers of a real service

That completes **Part 1: Fundamentals**. Next is **[Module 6: Concurrency fundamentals](/learn/06-concurrency-fundamentals/)**.
