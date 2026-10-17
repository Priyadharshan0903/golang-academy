# jobqueue

A background job queue stored in PostgreSQL: the Module 13 project.

- **Enqueue** jobs with a JSON payload, a delay and a retry limit, optionally inside your own transaction (`EnqueueWith`), so a job exists only if the work that created it committed.
- **Workers** claim jobs with `SELECT ... FOR UPDATE SKIP LOCKED`, so many workers (in one process or many) never take the same job.
- **Retries** with exponential backoff and jitter; jobs that keep failing end up `failed` with their last error.
- **Crash recovery**: a claimed job is locked for a visibility timeout; if its worker dies, another worker picks it up when the lock expires.
- **Graceful shutdown**: cancelling `Run` stops claiming new jobs and waits for running ones.

Delivery is **at least once**: a job can run twice if a worker dies after doing the work but before recording it. Handlers must be idempotent.

This directory is its own Go module (it depends on pgx).

## Run the tests

The tests need PostgreSQL and are skipped without it:

```bash
docker run -d --name gopath-pg -e POSTGRES_PASSWORD=pg -p 5432:5432 postgres:16
export TEST_DATABASE_URL='postgres://postgres:pg@localhost:5432/postgres?sslmode=disable'
go mod tidy
go test -race ./...
```

## Run the demo

```bash
DATABASE_URL="$TEST_DATABASE_URL" go run ./cmd/demo
```

It enqueues 20 fake emails and processes them with 4 workers; every fifth job fails once and is retried.
