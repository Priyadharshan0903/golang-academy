---
title: "13.2 pgx and PostgreSQL"
linkTitle: "13.2 pgx"
weight: 2
description: "Use pgx's native API: pgxpool, scanning rows into structs, PostgreSQL arrays, handling constraint violations by error code, and bulk inserts with batches and COPY, measured."
goVersion: "1.27"
---

## What you will learn

- When to use pgx's native API instead of `database/sql`
- Connecting with `pgxpool`
- Scanning rows into structs with `pgx.CollectRows`, and PostgreSQL types like arrays
- Handling errors by **SQLSTATE code**, such as unique violations
- Bulk inserts: one by one, batched, and `COPY`, with measurements

## database/sql or native pgx?

pgx offers two APIs. Through `database/sql` (lesson 13.1), your code is portable between databases and works with tools built on that interface. Through its **native API**, you get PostgreSQL-specific features and better performance: native types (arrays, JSONB, ranges, UUIDs), batches, `COPY`, `LISTEN/NOTIFY`, and helpers that scan rows into structs.

If you know you are on PostgreSQL, and most Go services are, the native API is the better default. The shortener and job queue projects use it.

## Connecting

{{< example file="examples/thirdparty/db/pgxdemo/pgxdemo.go" region="pool" >}}

`pgxpool.Pool` plays the same role as `*sql.DB`: a concurrency-safe pool, created once. Settings can also come from the connection URL (`?pool_max_conns=10`), which is convenient when the URL is the only configuration a platform passes.

## Rows into structs

{{< example file="examples/thirdparty/db/pgxdemo/pgxdemo.go" region="rows" >}}

`pgx.CollectRows(rows, pgx.RowToStructByName[Product])` reads every row into a `Product`, matching columns to fields by `db` tag (or by field name), and closes the rows. There are variants for positional mapping (`RowToStructByPos`), single values (`RowTo[string]`), and exactly one row (`CollectOneRow`, which returns `pgx.ErrNoRows` if there is none).

`Tags []string` maps to a PostgreSQL `text[]` column in both directions, and `$1 = ANY(tags)` searches it. One trap the test hit while this lesson was written: a **nil** Go slice is sent as SQL `NULL`, not as an empty array, and the `NOT NULL` column rejected it. `Insert` therefore sends `[]string{}` for a nil slice.

## Errors by code

PostgreSQL reports every error with a five-character **SQLSTATE** code. pgx exposes it as `*pgconn.PgError`:

{{< example file="examples/thirdparty/db/pgxdemo/pgxdemo.go" region="errors" >}}

```text
duplicate: insert "GO-MUG": duplicate SKU
check violation: insert "NEG": ERROR: new row for relation "products" violates check constraint "products_price_cents_check" (SQLSTATE 23514)
```

Codes worth knowing:

| Code | Name | Typical meaning |
|---|---|---|
| `23505` | unique_violation | duplicate key: map it to "already exists" (409) |
| `23503` | foreign_key_violation | referenced row missing |
| `23514` | check_violation | a `CHECK` constraint failed |
| `40001` | serialization_failure | retry the transaction (lesson 13.4) |
| `40P01` | deadlock_detected | retry, and fix the lock order (lesson 13.4) |
| `57014` | query_canceled | the context was cancelled or timed out |

Let the **database enforce** invariants (unique keys, foreign keys, checks) and translate its errors, rather than checking first in Go: "check, then insert" is a race between two requests, while a unique constraint is atomic.

## Bulk inserts

Inserting 5000 rows three ways, against PostgreSQL 16 on the same machine:

{{< example file="examples/thirdparty/db/pgxdemo/pgxdemo.go" region="bulk" >}}

{{< example file="examples/thirdparty/db/pgxdemo/pgxdemo_test.go" region="bulktest" run="cd examples/thirdparty && go test -run TestBulkInsert -v ./db/pgxdemo" >}}

```text
one by one 5000 rows in 2.379s
batch      5000 rows in 80ms
copy       5000 rows in 23ms
```

One by one is **30 times** slower than a batch, and **100 times** slower than `COPY`, even with the database on the same machine. Two costs add up:

- **Round trips.** Each `Exec` waits for its answer before the next is sent. Over a real network, with a millisecond or so per round trip, this grows much worse than here.
- **Commits.** Outside a transaction, each statement is its own transaction, and each commit waits for the write-ahead log to reach disk.

A **batch** queues all statements and sends them together, and pgx runs them in a single implicit transaction. **`COPY`** is PostgreSQL's bulk-loading protocol: it streams rows in a compact format, without parsing an `INSERT` per row. Use batches for many statements of any kind, and `CopyFrom` for loading many rows into one table.

## Common mistakes

### Checking for duplicates before inserting

Two requests can both pass the check. Use a unique constraint and handle `23505`.

### Nil slices for NOT NULL array columns

They become NULL. Send an empty slice.

### Inserting rows in a loop

Use `SendBatch` or `CopyFrom` for anything more than a handful.

### Matching error message text

Use `pgErr.Code` (and `pgErr.ConstraintName` when a table has several constraints).

## Key takeaways

- On PostgreSQL, prefer pgx's native API and `pgxpool`.
- `pgx.CollectRows` with `RowToStructByName` scans rows into structs; arrays and other PostgreSQL types map to Go types.
- Classify errors with `errors.As` into `*pgconn.PgError` and its SQLSTATE `Code`; let constraints enforce invariants.
- For bulk work, batches were 30x and `COPY` 100x faster than one statement at a time, here.

## Further reading

- [pgx documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [PostgreSQL error codes](https://www.postgresql.org/docs/current/errcodes-appendix.html)
- [PostgreSQL: COPY](https://www.postgresql.org/docs/current/sql-copy.html)
