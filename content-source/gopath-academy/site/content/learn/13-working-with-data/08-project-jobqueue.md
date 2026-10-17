---
title: "13.8 Project: a PostgreSQL job queue"
linkTitle: "13.8 Project: job queue"
weight: 8
description: "Build a background job queue on PostgreSQL: transactional enqueue, workers that claim jobs with FOR UPDATE SKIP LOCKED, retries with exponential backoff, crash recovery through lock expiry, and graceful shutdown, all tested against a real database."
goVersion: "1.27"
---

## The project

Sending an email, resizing an image, calling a slow partner API: work that should not happen while an HTTP client waits belongs in a **background job queue**. Producers record jobs; workers process them later, retrying failures.

You can use a dedicated broker (RabbitMQ, SQS, Kafka) for this, but many teams start with the PostgreSQL database they already run, and some never leave. This project builds such a queue in about 300 lines, comments included, using the patterns of this module and of Module 7. The code is in [`projects/jobqueue`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/jobqueue).

What it provides:

- Enqueue with a JSON payload, delay and retry limit, **optionally inside your own transaction**.
- Many workers, in one process or many, each job claimed by exactly one of them.
- Retries with exponential backoff and jitter, then `failed` with the last error.
- **Crash recovery**: if a worker dies mid-job, another picks the job up when its lock expires.
- Graceful shutdown that lets running jobs finish.

## The table

{{< example file="projects/jobqueue/schema.sql" >}}

A job moves through `queued` → `running` → `done`, or back to `queued` for a retry, or to `failed`. The partial index covers only jobs that can still run, so it stays small however many finished jobs the table keeps.

## Enqueue, in your transaction

{{< example file="projects/jobqueue/jobqueue.go" region="enqueue" >}}

`EnqueueWith` accepts anything with a `QueryRow` method: the pool, a connection, or a **transaction**. This is the main reason to keep a queue in your own database. Consider "create an order, then queue the confirmation email": with an external broker, the order can commit while the enqueue fails, or the email can be sent for an order that was then rolled back. With the job in the same transaction as the order, both happen or neither does:

{{< example file="projects/jobqueue/jobqueue_test.go" region="txenqueue" >}}

## Claiming a job: SKIP LOCKED

{{< example file="projects/jobqueue/jobqueue.go" region="claim" >}}

The inner `SELECT` finds the oldest runnable job and locks it with `FOR UPDATE`. On its own, that would make concurrent workers **wait** for each other's locks, processing jobs one at a time. **`SKIP LOCKED`** tells PostgreSQL to skip rows that another transaction has locked and take the next one. Eight workers claiming at once get eight different jobs, without blocking.

The outer `UPDATE` marks the job `running`, counts the attempt, and sets `locked_until`: the job's **visibility timeout**. The same `WHERE` clause also accepts `running` jobs whose lock has expired. That single condition is the whole crash-recovery mechanism.

{{< example file="projects/jobqueue/jobqueue_test.go" region="exactlyonce" run="cd projects/jobqueue && go test -run TestEveryJobProcessedOnce -v ." >}}

```text
8 workers processed 200 jobs in 159ms
```

Every job processed, each exactly once.

## Finishing, retrying, giving up

{{< example file="projects/jobqueue/jobqueue.go" region="finish" >}}

Both updates include `AND attempts = $2`. Suppose a job takes longer than its visibility timeout: its lock expires, another worker claims it (attempts goes from 1 to 2), and then the first worker finishes. Without the check, the first worker would mark a job done that the second is still running. With it, the stale update matches no row. This is **optimistic concurrency control**, with the attempt counter as the version number.

## The worker

{{< example file="projects/jobqueue/worker.go" region="run" >}}

The structure is a worker pool from lesson 7.2, with three production details:

- **Jitter** on the poll interval, so idle workers spread out instead of hitting the database in lockstep.
- **Panics become failures** (`safeCall`), so one bad payload cannot crash the process.
- **Two contexts.** The `Run` context means "stop taking new work". Each job gets its own context with the visibility timeout as its deadline: shutting down does not interrupt a running job, and a job cannot run past its lock without its handler being told.

## Crash recovery, tested

{{< example file="projects/jobqueue/jobqueue_test.go" region="reclaim" run="cd projects/jobqueue && go test -run TestReclaimAfterCrash -v ." >}}

```text
job 1 reclaimed after 220ms, on attempt 2
```

The test claims a job and abandons it, like a worker killed by the OOM killer. A real worker picks it up once the 200 ms lock expires.

This is also the honest limit of the design: delivery is **at least once**. A worker that sends the email and then dies before marking the job done will cause a second email. Every queue that survives crashes has this property; the remedy is **idempotent handlers**, for example recording "email for order 42 sent" in the same database and checking it first.

## Graceful shutdown

{{< example file="projects/jobqueue/jobqueue_test.go" region="shutdown" >}}

```text
Run returned 300ms after cancel
```

`Run` returns only after the running job completes. In Kubernetes, set `terminationGracePeriodSeconds` longer than your longest job, or a job will be killed mid-way and retried after its lock expires, which is safe, but slower.

## Try the demo

```bash
cd projects/jobqueue
DATABASE_URL='postgres://postgres:pg@localhost:5432/postgres?sslmode=disable' go run ./cmd/demo
```

It enqueues 20 fake emails for 4 workers; every fifth job fails once and is retried with backoff.

## When to use a dedicated broker instead

A PostgreSQL queue is a good fit for up to hundreds of jobs per second, when you value transactional enqueue and one less system to operate. Consider a broker when you need much higher throughput, fan-out of one message to many consumers, or long retention of event streams. Mature Go libraries built on the same Postgres ideas exist too, such as River (`github.com/riverqueue/river`).

## Your turn

1. Replace polling with **`LISTEN/NOTIFY`**: `Enqueue` runs `NOTIFY jobs`, and idle workers wait on `conn.WaitForNotification` with the poll interval as a fallback.
2. Add a **dead-letter** view: an admin endpoint listing `failed` jobs with their last error, and a way to re-queue one.
3. Export metrics (lesson 12.5): jobs processed and failed by queue, processing duration, and the queue depth from `Counts`.
4. Add **unique jobs**: a nullable `unique_key` column with a partial unique index on queued and running jobs, so "send welcome email to user 42" cannot be enqueued twice.
