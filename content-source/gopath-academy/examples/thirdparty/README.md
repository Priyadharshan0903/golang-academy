# Third party examples

Examples that use packages outside the standard library live in this
separate module, so the main course module stays dependency-free.

```bash
cd examples/thirdparty
go mod tidy      # downloads dependencies and writes go.sum the first time
go test ./...
```

CI runs the same commands. Dependencies are pinned in `go.mod`.

## Databases (Module 13)

Tests that need PostgreSQL or Redis are **skipped** unless these
variables are set. Each database test creates its own schema and drops
it afterwards, so any scratch database works.

```bash
docker run -d --name gopath-pg -e POSTGRES_PASSWORD=pg -p 5432:5432 postgres:16
docker run -d --name gopath-redis -p 6379:6379 redis:7

export TEST_DATABASE_URL='postgres://postgres:pg@localhost:5432/postgres?sslmode=disable'
export TEST_REDIS_ADDR='localhost:6379'
go test ./db/... ./cache/...
```

| Package | Lesson |
|---|---|
| `observability` | 12.5 Prometheus metrics and OpenTelemetry tracing |
| `db/sqlbasics` | 13.1 `database/sql`, pooling, SQL injection |
| `db/pgxdemo` | 13.2 pgx: structs, errors, batch and COPY |
| `db/txdemo` | 13.4 transactions, lost updates, deadlocks, cancellation |
| `cache/rediscache` | 13.5 cache-aside with Redis and request coalescing |
