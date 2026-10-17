---
title: "13.7 Project milestone: the shortener on PostgreSQL"
linkTitle: "13.7 Project: shortener on Postgres"
weight: 7
description: "Milestone 3 of the URL shortener: a PostgreSQL store with embedded migrations, an optional Redis cache, and one conformance test suite that every store implementation must pass."
goVersion: "1.27"
---

## The milestone

| Milestone | After module | Adds | Status |
|---|---|---|---|
| 1. Core | 5. Errors | Domain types, errors, validation, `Service`, in-memory store | done |
| 2. HTTP API | 12. HTTP services | JSON API, redirects, middleware, metrics, shutdown | done |
| **3. Postgres** (this lesson) | 13. Working with data | PostgreSQL store, migrations, Redis cache | done |
| 4. Deploy | 15. Cloud native | Container image and Kubernetes manifests | next |

Links now survive restarts and can be shared by many replicas. The interesting part is what did **not** change: the `link` package and the whole HTTP layer are untouched. The `Store` interface from milestone 1 was designed for this moment.

```text
internal/link/linktest/   a conformance suite every Store must pass      (new)
internal/postgres/        Store on pgx, migrations embedded in the binary (new)
internal/cache/           a caching Store wrapping another Store          (new)
internal/pgtest/          a fresh schema per test                         (new)
cmd/shortener/            chooses the store from the configuration        (changed)
```

## Run it with a database

```bash
docker run -d --name gopath-pg -e POSTGRES_PASSWORD=pg -p 5432:5432 postgres:16
docker run -d --name gopath-redis -p 6379:6379 redis:7

cd projects/shortener
DATABASE_URL='postgres://postgres:pg@localhost:5432/postgres?sslmode=disable' \
REDIS_ADDR=localhost:6379 LOG_FORMAT=text go run ./cmd/shortener
```

On start-up it applies migrations and logs `migrations applied=[0001_create_links.sql]`; the second start logs `applied=[]`. Without `DATABASE_URL` it still runs in memory, with a warning.

## The Postgres store

{{< example file="projects/shortener/internal/postgres/store.go" >}}

Two translations make it a correct `link.Store`:

- **Unique violation (23505) becomes `link.ErrCodeTaken`.** The primary key on `code` makes the database decide, atomically, which of two concurrent requests for the same custom code wins. The service's retry loop for random codes (milestone 1) keeps working unchanged, because it only looks for `ErrCodeTaken`.
- **`pgx.ErrNoRows` becomes `link.ErrNotFound`**, which the HTTP layer already maps to 404.

And one normalization: pgx returns `timestamptz` values in the local time zone, so `Get` converts to UTC, as the in-memory store returns them.

Migrations are the ones from lesson 13.3, embedded in the binary and applied under an advisory lock.

## One test suite for every store

How do you know the Postgres store behaves like the memory store that all the HTTP tests used? Write the expectations once, as a **conformance suite**, and run it against every implementation:

{{< example file="projects/shortener/internal/link/linktest/linktest.go" >}}

Each implementation needs only a few lines:

```go
func TestStoreContract(t *testing.T) {
	linktest.Run(t, func(t *testing.T) link.Store {
		pool := pgtest.Pool(t)          // a fresh, empty schema
		if _, err := postgres.Migrate(context.Background(), pool); err != nil {
			t.Fatal(err)
		}
		return postgres.NewStore(pool)
	})
}
```

The suite runs for `MemoryStore`, the Postgres store, and the cached store. It pins down details that are easy to get subtly different, such as returning `CreatedAt` in UTC even though pgx returns local time, and keeping only microseconds as PostgreSQL does. `ConcurrentCreateSameCode` checks that 20 simultaneous creates of one code produce exactly one success and 19 `ErrCodeTaken`, in memory (a mutex) and in Postgres (the primary key) alike.

This pattern, an exported test suite in a `...test` package, is how the standard library tests interfaces too: `testing/fstest.TestFS` checks any `fs.FS` implementation, and `golang.org/x/net/nettest` does the same for `net.Conn`.

## The cache is just another Store

{{< example file="projects/shortener/internal/cache/cache.go" >}}

`cache.Store` wraps any `link.Store` and is itself a `link.Store`: the decorator pattern, through a one-method-plus-one interface. Links never change after creation, so there is no invalidation, only a TTL to bound memory. Misses are not cached, so a link created a moment after a failed lookup is found immediately. Redis errors fall back to the database (lesson 13.5).

## Choosing the store

{{< example file="projects/shortener/cmd/shortener/main.go" region="store" >}}

`openStore` builds the chain from configuration: memory or Postgres, optionally wrapped in the cache, and returns one function that closes everything in reverse order (`slices.Backward`, Go 1.23). The database URL is a secret, so it can come from `DATABASE_URL_FILE` (lesson 12.4) and is never logged. Note the different attitude to failures: an unreachable **database** stops start-up (the service cannot work), an unreachable **Redis** only produces a warning (the service works, just slower).

## Tests

```bash
export TEST_DATABASE_URL='postgres://postgres:pg@localhost:5432/postgres?sslmode=disable'
export TEST_REDIS_ADDR=localhost:6379
go test -race ./...
```

Without those variables, the database tests are skipped and the rest still run, so the project stays easy to work on without Docker. CI runs everything, with PostgreSQL 16 and Redis 7 as service containers. `TestRunWithPostgresAndRedis` runs the whole program, as `TestRun` does in memory, against both real servers.

## Your turn

1. Milestone 4 adds `Delete` to the `Store` interface. Before reading it, try it yourself: add the case to the conformance suite **first**, then implement it in each store. What must the cache do on delete, and in which order?
2. Count visits per link. Increment a column with `UPDATE links SET visits = visits + 1` (atomic, lesson 13.4); add the column with a new migration. Does doing this on every redirect undo the benefit of the cache? Consider enqueueing increments in batches (lesson 13.8).
3. Make readiness check the database with a short timeout. Then argue whether it should: what happens to all replicas when the database has a five second hiccup?
