# shortener

A URL shortener built across several modules of the course.

| Milestone | Module | What is added | Status |
|---|---|---|---|
| 1. Core | 5 (Errors) | `internal/link`: domain type, sentinel errors, validation, `Service`, in-memory `Store` | done |
| 2. HTTP API | 12 (HTTP services) | `internal/httpapi`, `internal/config`, `cmd/shortener`: JSON API, redirects, middleware, metrics, graceful shutdown | done |
| 3. Postgres | 13 (Working with data) | `internal/postgres` (pgx store, embedded migrations), `internal/cache` (Redis), `internal/link/linktest` (store conformance suite) | done |
| 4. Kubernetes | 15 (Cloud native) | `Dockerfile`, `k8s/` (Kustomize base and local overlay), `DELETE` endpoint for the ShortLink operator | done |

This directory is its own Go module, because it uses third party packages (pgx, go-redis, the Prometheus client, `golang.org/x/time/rate` and `golang.org/x/sync`). The rest of the course has no dependencies.

## Layout

```text
cmd/shortener/      main: wiring, servers, graceful shutdown
internal/link/      the domain: Link, Service, Store, errors (milestone 1)
internal/link/linktest/  conformance tests every Store must pass
internal/httpapi/   routes, JSON, error mapping, middleware, metrics, rate limiting
internal/config/    configuration from environment variables
internal/postgres/  PostgreSQL Store and migrations
internal/cache/     Redis cache in front of any Store
internal/pgtest/    test helper: a fresh schema per test
```

## Run it

```bash
cd projects/shortener
go mod tidy          # first time only: downloads dependencies
go test ./...
LOG_FORMAT=text go run ./cmd/shortener
```

Without `DATABASE_URL`, links are kept in memory. With PostgreSQL and Redis:

```bash
docker run -d --name gopath-pg -e POSTGRES_PASSWORD=pg -p 5432:5432 postgres:16
docker run -d --name gopath-redis -p 6379:6379 redis:7
export DATABASE_URL='postgres://postgres:pg@localhost:5432/postgres?sslmode=disable'
REDIS_ADDR=localhost:6379 LOG_FORMAT=text go run ./cmd/shortener
```

Database tests are skipped unless `TEST_DATABASE_URL` (and, for the cache, `TEST_REDIS_ADDR`) are set:

```bash
export TEST_DATABASE_URL="$DATABASE_URL" TEST_REDIS_ADDR=localhost:6379
go test -race ./...
```

With the service running, in another terminal:

```bash
curl -i -X POST localhost:8080/api/links -H 'Content-Type: application/json' \
  -d '{"url": "https://go.dev/doc", "code": "godoc"}'
curl -i localhost:8080/godoc            # 302 to https://go.dev/doc
curl -s localhost:9090/metrics | grep shortener_
```

## Kubernetes

See lesson 15.7. In short: build the image, load it into a local cluster, then `kubectl apply -k k8s/overlays/local`. The ShortLink operator in `projects/shortlink-operator` manages links declaratively.

## API

| Method and path | Success | Errors |
|---|---|---|
| `POST /api/links` with `{"url": "...", "code": "optional"}` | 201, link JSON, `Location` header | 400 bad JSON, 409 code taken, 415 wrong content type, 422 invalid URL or code, 429 rate limited |
| `GET /api/links/{code}` | 200, link JSON | 404, 422, 429 |
| `DELETE /api/links/{code}` | 204 | 404, 422, 429 |
| `GET /{code}` | 302 redirect to the target | 404, 422, 429 |
| `GET /healthz` | 200 while the process runs (liveness) | |
| `GET /readyz` | 200 while accepting traffic (readiness) | 503 before start-up completes and during shutdown |

Errors are `application/problem+json` (RFC 9457) and include the request ID. Server errors are logged in full but never shown to clients.

The admin port serves `GET /metrics` (Prometheus) and `/debug/pprof/`. Never expose it publicly.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `ADDR` | `:8080` | API listen address |
| `ADMIN_ADDR` | `:9090` | metrics and pprof listen address |
| `BASE_URL` | `http://localhost:8080` | public URL, used in `short_url` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |
| `LOG_FORMAT` | `json` | `json` or `text` |
| `RATE_LIMIT` | `5` | requests per second per client |
| `RATE_BURST` | `10` | burst size per client |
| `CLIENT_IP_HEADER` | (none) | header with the client IP, such as `X-Real-IP`; only set this behind a proxy that always sets it |
| `SHUTDOWN_DELAY` | `0s` | time to keep serving while reporting not ready, before shutting down (use about 5s in Kubernetes) |
| `SHUTDOWN_TIMEOUT` | `10s` | time allowed for in-flight requests during shutdown |
| `DATABASE_URL` or `DATABASE_URL_FILE` | (none: in memory) | PostgreSQL connection URL, or a file containing it |
| `MIGRATE` | `true` | apply migrations at start-up |
| `REDIS_ADDR` | (none: no cache) | Redis address, such as `localhost:6379` |
| `CACHE_TTL` | `1h` | how long links stay in the cache |
