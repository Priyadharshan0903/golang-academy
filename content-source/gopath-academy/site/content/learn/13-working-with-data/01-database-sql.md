---
title: "13.1 database/sql and connection pooling"
linkTitle: "13.1 database/sql"
weight: 1
description: "How database/sql works: drivers, the connection pool and its settings, queries and rows, NULL values, sql.ErrNoRows, and why parameters prevent SQL injection, with a measured pool limit."
goVersion: "1.27"
---

## What you will learn

- How `database/sql` and drivers fit together
- What `*sql.DB` really is: a **pool**, and how to size it
- Querying one row, many rows, and executing statements, correctly
- NULLs, `sql.ErrNoRows`, and closing rows
- How parameters prevent SQL injection, demonstrated

## The idea

`database/sql` is the standard library's interface to SQL databases. It contains **no driver**: it defines the API your code uses, and a driver package implements it for a specific database. You import the driver for its side effect of registering itself:

```go
import _ "github.com/jackc/pgx/v5/stdlib" // registers the driver name "pgx"
```

For PostgreSQL, **pgx** is the most widely used and most complete driver; its `stdlib` package provides the `database/sql` driver. (You will also meet the older `lib/pq` in existing code.) Lesson 13.2 shows pgx's own, richer API.

## sql.DB is a pool

{{< example file="examples/thirdparty/db/sqlbasics/users.go" region="open" >}}

`sql.Open` does not connect; it only checks its arguments. The `*sql.DB` it returns is a **pool of connections**, safe for concurrent use. Create one at start-up, share it everywhere, and close it at shutdown. Opening a `sql.DB` per request is a classic, expensive mistake.

Each query borrows a connection from the pool and returns it when done. The settings:

| Setting | Default | What to set |
|---|---|---|
| `SetMaxOpenConns` | unlimited | Always set it. The database has a connection limit (PostgreSQL's default is 100) shared by every instance of every service. 10 replicas × 25 connections is already 250. |
| `SetMaxIdleConns` | 2 | Usually equal to MaxOpenConns, so bursts do not open and close connections constantly. |
| `SetConnMaxLifetime` | forever | Minutes to an hour, so connections are recycled (helps with load balancers and failovers). |
| `SetConnMaxIdleTime` | forever | A few minutes, so idle capacity is released. |

### What happens when the pool is full

When all connections are busy, the next query **waits** for one to be returned. This test allows 2 connections and runs 6 queries that each take 200 ms:

{{< example file="examples/thirdparty/db/sqlbasics/users_test.go" region="pooltest" run="cd examples/thirdparty && go test -run TestPoolLimit -v ./db/sqlbasics" >}}

```text
6 queries of 200ms with 2 connections took 610ms; waited 4 times for 1.22s in total
```

Three rounds of two queries: 600 ms. `db.Stats()` shows the waiting directly. Export these stats as metrics (lesson 12.5): a growing `WaitCount` means the pool is too small, or queries are too slow, and requests are queuing. The wait respects the query's context, so a request deadline also bounds the time spent waiting for a connection.

## Queries

{{< example file="examples/thirdparty/db/sqlbasics/users.go" region="crud" >}}

The three methods, always the `Context` variants so that cancellation and deadlines apply:

- **`QueryRowContext(...).Scan(...)`** for exactly one row. If there is none, `Scan` returns **`sql.ErrNoRows`**; translate it into your own "not found" error, so callers do not depend on `database/sql`.
- **`QueryContext`** for many rows. Three rules: `defer rows.Close()` immediately (an unclosed `Rows` keeps its connection out of the pool, a pool leak), check `rows.Scan`'s error, and check **`rows.Err()`** after the loop, because a network error mid-result just ends the loop.
- **`ExecContext`** for statements that return no rows. Its `Result` has `RowsAffected`.

**NULL** needs a type that can represent it. `string` cannot be NULL, so scanning a NULL into it fails. Use `sql.Null[string]` (Go 1.22), which has `V` and `Valid`, or a pointer (`*string`). Better still, avoid nullable columns that do not need to be.

PostgreSQL uses `$1, $2, ...` placeholders; MySQL and SQLite use `?`. `RETURNING id` gets the generated ID in the same round trip.

## SQL injection

Never build SQL by formatting values into the string:

{{< example file="examples/thirdparty/db/sqlbasics/users.go" region="injection" >}}

{{< example file="examples/thirdparty/db/sqlbasics/users_test.go" region="injectiontest" >}}

```text
unsafe query matched 2 users, safe query matched 0
```

The input `nobody' OR '1'='1` closes the string literal and adds a condition that is always true, so the unsafe query matches every user. With a parameter, the value is sent **separately** from the SQL text, and the database never parses it as SQL: it simply looks for a user with that strange name. Parameters work for values only; if you must vary an identifier (a column to sort by), choose it from a fixed allowlist in code.

## Common mistakes

### A new sql.DB per request

Create one pool at start-up and share it.

### No MaxOpenConns

Under load, every instance opens connections until the database refuses them, failing every service that uses it.

### Forgetting rows.Close or rows.Err

The first leaks connections until the pool is exhausted; the second silently returns partial results.

### Formatting values into SQL

Use parameters, always.

## Key takeaways

- `database/sql` is an interface; import a driver such as `github.com/jackc/pgx/v5/stdlib`.
- `*sql.DB` is a concurrency-safe pool: one per database per program, with `SetMaxOpenConns` and friends set deliberately.
- When the pool is full, queries wait; `db.Stats()` shows it.
- Use the `Context` methods; map `sql.ErrNoRows` to your own error; always `rows.Close()` and check `rows.Err()`.
- Pass values as parameters (`$1`), never with string formatting.

## Further reading

- [Go database/sql tutorial](https://go.dev/doc/database/) on go.dev
- [Managing connections](https://go.dev/doc/database/manage-connections)
- [pgx documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
