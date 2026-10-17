---
title: "13.4 Transactions and context"
linkTitle: "13.4 Transactions"
weight: 4
description: "Use transactions correctly from Go: a WithTx helper, the lost update problem and its fixes, row locks with FOR UPDATE, deadlocks and lock ordering, isolation levels and retries, and cancelling queries with context, all demonstrated against PostgreSQL."
goVersion: "1.27"
---

## What you will learn

- A safe pattern for transactions in Go: commit on success, roll back on everything else
- The **lost update** problem, reproduced, and two ways to fix it
- Row locks with `SELECT ... FOR UPDATE`
- **Deadlocks**, reproduced, and the lock-ordering fix
- Isolation levels and retrying serialization failures
- How a context deadline cancels a running query

## The pattern

A transaction groups statements so they succeed or fail together. In Go, the shape is always the same:

{{< example file="examples/thirdparty/db/txdemo/txdemo.go" region="withtx" >}}

- **`defer tx.Rollback(...)`** runs on every exit path: an error, an early return, a panic. After a successful `Commit` it does nothing, so it is safe to defer unconditionally.
- **`context.WithoutCancel`** for the rollback: if the request was cancelled, we still want the rollback to reach the database.
- The callback gets a `pgx.Tx` and runs every statement through it. A statement accidentally run on the pool instead of `tx` executes **outside** the transaction, a subtle bug.

pgx ships the same idea as `pgx.BeginFunc`, which the migrator in lesson 13.3 uses. The test confirms the rollback:

{{< example file="examples/thirdparty/db/txdemo/txdemo_test.go" region="rollbacktest" >}}

## The lost update

A transaction alone does not make read-then-write logic safe. This withdrawal reads the balance, computes, and writes the result:

{{< example file="examples/thirdparty/db/txdemo/txdemo.go" region="lost" >}}

Two withdrawals of 10 from a balance of 100, at the same time:

{{< example file="examples/thirdparty/db/txdemo/txdemo_test.go" region="losttest" run="cd examples/thirdparty && go test -run TestLostUpdate -v ./db/txdemo" >}}

```text
racy:   two withdrawals of 10 from 100 left 90
locked: two withdrawals of 10 from 100 left 80
```

In the racy version, both transactions read 100 and both write 90: one withdrawal vanished. PostgreSQL's default isolation level, **READ COMMITTED**, does not prevent this, and neither does a transaction by itself. (The test widens the gap between read and write with a pause, so the race happens every time instead of occasionally in production.)

Two fixes:

1. **Lock the row** with `SELECT ... FOR UPDATE`. The second transaction blocks at its `SELECT` until the first commits, then reads 90. That is `WithdrawLocked`, which leaves 80.
2. **Do the arithmetic in the database**: `UPDATE accounts SET balance = balance - $1 WHERE id = $2`. A single statement reads and writes atomically, with no window for a race. This is simpler and faster whenever the logic fits in SQL; the `Transfer` function below uses it, with `AND balance >= $1` to refuse overdrafts.

## Deadlocks

`TransferNaive` moves money with two atomic `UPDATE`s in one transaction. Now run transfers in **both** directions at once:

{{< example file="examples/thirdparty/db/txdemo/txdemo.go" region="transfer" >}}

{{< example file="examples/thirdparty/db/txdemo/txdemo_test.go" region="deadlocktest" run="cd examples/thirdparty && go test -run TestDeadlock -v ./db/txdemo" >}}

```text
naive transfers, errors by code: map[40P01:6 insufficient funds:4] (40P01 is deadlock_detected)
ordered locking, errors by code: map[]
```

A transfer from alice to bob locks alice's row, then bob's. A transfer from bob to alice locks bob's, then alice's. When they interleave, each holds the lock the other needs, forever. PostgreSQL detects the cycle (after `deadlock_timeout`, one second by default) and aborts one of the transactions with SQLSTATE **40P01**. The total money stays correct, because the aborted transaction rolls back, but requests fail.

The fix is the same as for code that takes several `sync.Mutex` locks (lesson 6.4): **acquire locks in a consistent order**. `Transfer` first locks both rows with `SELECT ... ORDER BY id FOR UPDATE`, so every transfer locks alice before bob regardless of direction, and no cycle can form. Zero deadlocks.

## Isolation levels and retries

PostgreSQL offers three isolation levels in practice:

| Level | Guarantees | Cost |
|---|---|---|
| READ COMMITTED (default) | each statement sees data committed before it started | lost updates possible, as shown |
| REPEATABLE READ | the whole transaction sees one snapshot; concurrent updates to the same row fail with **40001** | you must retry |
| SERIALIZABLE | the result equals some serial order of the transactions; conflicts fail with **40001** | you must retry, more often |

```go
tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
```

Higher levels replace "silently wrong" with "fails, please retry". Code using them needs a retry loop: run the whole transaction again, a few times with backoff, when the error code is `40001` (serialization_failure) or `40P01` (deadlock_detected). Never retry part of a transaction.

A practical approach: use READ COMMITTED with explicit locking (`FOR UPDATE`) or single-statement updates for the hot paths you understand, and SERIALIZABLE for complex invariants that are hard to protect by hand.

## Context cancels queries

Every pgx method takes a context. When the context ends, pgx asks the server to cancel the running query:

{{< example file="examples/thirdparty/db/txdemo/txdemo_test.go" region="canceltest" run="cd examples/thirdparty && go test -run TestContextCancelsQuery -v ./db/txdemo" >}}

```text
a 5 s query with a 100 ms deadline returned after 100ms: timeout: context deadline exceeded
```

This is why handlers pass `r.Context()` to every query: when a client disconnects or a deadline passes, the database stops working on a result nobody will read. The error wraps `context.DeadlineExceeded` (or `context.Canceled`), so `errors.Is` identifies it.

Keep transactions **short**. While a transaction is open it holds a connection and its locks; calling another service or waiting for user input inside one turns a slow dependency into blocked rows for everyone.

## Common mistakes

### Read, compute, write without a lock

Lost updates under concurrency. Use one `UPDATE ... SET x = x - $1`, or `FOR UPDATE`.

### Locking rows in different orders

Deadlocks. Lock in a consistent order, such as by primary key.

### Using the pool inside a transaction function

The statement runs outside the transaction. Use the `tx` you were given.

### Long transactions

Network calls inside a transaction hold locks and connections. Do them before or after.

### Retrying only the failed statement

After 40001 or 40P01 the whole transaction is aborted. Retry the whole function.

## Key takeaways

- Begin, `defer Rollback`, run everything on `tx`, `Commit`: or use a helper like `WithTx` or `pgx.BeginFunc`.
- READ COMMITTED allows lost updates: compute in a single `UPDATE`, or lock with `SELECT ... FOR UPDATE`.
- Lock rows in a consistent order to avoid deadlocks (40P01).
- REPEATABLE READ and SERIALIZABLE turn anomalies into 40001 errors, which require retrying the whole transaction.
- Contexts cancel running queries; pass the request context everywhere, and keep transactions short.

## Further reading

- [PostgreSQL: transaction isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
- [PostgreSQL: explicit locking](https://www.postgresql.org/docs/current/explicit-locking.html)
