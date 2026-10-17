---
title: "12.8 Project milestone: the shortener's HTTP API"
linkTitle: "12.8 Project: shortener API"
weight: 8
description: "Milestone 2 of the URL shortener: a production-shaped HTTP service with a JSON API, redirects, error mapping, request IDs, access logs, Prometheus metrics, per-client rate limiting, health checks and graceful shutdown, tested end to end."
goVersion: "1.27"
---

## The milestone

In milestone 1 (lesson 5.7) you built the shortener's core: the `link` package, with no knowledge of HTTP. Milestone 2 puts a service in front of it, using everything from this module. The core does not change at all; that is the payoff of keeping it independent.

| Milestone | After module | Adds | Status |
|---|---|---|---|
| 1. Core | 5. Errors | Domain types, errors, validation, `Service`, in-memory store | done |
| **2. HTTP API** (this lesson) | 12. HTTP services | JSON API, redirects, middleware, metrics, shutdown | done |
| 3. Postgres | 13. Working with data | A database-backed `Store`, migrations | next |
| 4. Deploy | 15. Cloud native | Container image and Kubernetes manifests | |

The code is in [`projects/shortener`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/shortener). The shortener is now **its own Go module**, because it uses two third party packages: the Prometheus client, and `golang.org/x/time/rate` for rate limiting (the production version of the token bucket you built in lesson 7.8).

```text
cmd/shortener/      main: wiring, the two servers, graceful shutdown
internal/link/      the domain from milestone 1, unchanged
internal/httpapi/   routes, JSON, error mapping, middleware, metrics, rate limiting
internal/config/    configuration from environment variables
```

Everything is under `internal/`, so no other module can import it: the only public interface of this program is HTTP.

## Try it

```bash
cd projects/shortener
go mod tidy                      # first time only
LOG_FORMAT=text go run ./cmd/shortener
```

```bash
$ curl -s -X POST localhost:8080/api/links -H 'Content-Type: application/json' \
    -d '{"url": "https://go.dev/doc", "code": "godoc"}'
{"code":"godoc","url":"https://go.dev/doc","short_url":"http://localhost:8080/godoc","created_at":"..."}

$ curl -si localhost:8080/godoc | grep -E '^(HTTP|Location)'
HTTP/1.1 302 Found
Location: https://go.dev/doc

$ curl -s -X POST localhost:8080/api/links -H 'Content-Type: application/json' \
    -d '{"url": "https://go.dev", "code": "godoc"}'
{"title":"Conflict","status":409,"detail":"create \"godoc\": code already in use","request_id":"..."}

$ curl -s localhost:9090/metrics | grep 'shortener_http_requests_total{'
shortener_http_requests_total{code="201",route="POST /api/links"} 1
shortener_http_requests_total{code="302",route="GET /{code}"} 1
shortener_http_requests_total{code="409",route="POST /api/links"} 1
```

## The routes and the middleware stack

{{< example file="projects/shortener/internal/httpapi/api.go" region="new" >}}

Things to notice:

- **Rate limiting wraps individual routes** (lesson 12.2), so health checks are never limited. Because it runs inside the mux, rate limited requests still get their route label in metrics.
- **Recovery sits inside logging and metrics**, so a panicking request is logged and counted as a 500 like any other.
- **`GET /{code}` and `GET /healthz` do not conflict**: `/healthz` is more specific (lesson 12.1). It does mean `healthz`, `readyz` and `api` can never be short codes; a real service would reserve them in validation.

## The handlers

{{< example file="projects/shortener/internal/httpapi/api.go" region="handlers" >}}

Each handler is a straight line: decode, call the service, respond or hand the error to `writeError`. The redirect uses **302 Found**, not 301: browsers cache permanent redirects indefinitely, so a 301 could never be counted or changed again.

## Mapping errors to status codes

Milestone 1 promised that "the HTTP layer maps each sentinel error to a status code". Here it is, in one place:

{{< example file="projects/shortener/internal/httpapi/errors.go" region="errors" >}}

Because the service wraps errors with `%w`, `errors.Is` finds `link.ErrCodeTaken` inside `create "godoc": code already in use`. Client errors show their detail; server errors are **logged in full and hidden from the client**, who gets a generic message and the request ID to quote in a bug report. `TestServerErrorsAreNotLeaked` checks both halves with a store that fails like a broken database.

## Wiring it together

`main` is three lines: create a context that is cancelled on SIGINT or SIGTERM, call `run`, and exit non-zero if it fails. Everything else is in `run`:

{{< example file="projects/shortener/cmd/shortener/main.go" region="wiring" >}}

`run` takes `getenv` and `stdout` as parameters and stops when its context is cancelled. That makes the **whole program testable**: `TestRun` starts it on free ports (`ADDR=127.0.0.1:0`), finds the actual addresses in the JSON log, creates a link over real HTTP, follows the redirect, scrapes the admin port's metrics, then cancels the context, exactly what a SIGTERM does, and checks that `run` returns `nil`.

Two servers: the **API** on `:8080` and the **admin** server on `:9090` with `/metrics` and pprof, never exposed publicly (lesson 12.5). The pprof handlers are registered on the admin mux explicitly instead of importing `net/http/pprof` for its side effect on the default mux.

## Graceful shutdown, Kubernetes style

{{< example file="projects/shortener/cmd/shortener/main.go" region="shutdown" >}}

Lesson 7.7 introduced `Server.Shutdown`. In Kubernetes there is one more subtlety: when a pod is terminated, removing it from the Service's endpoints and sending it SIGTERM happen **at the same time**. For a few seconds, other components may still route new requests to a pod that has started shutting down. So the shortener:

1. turns `/readyz` to 503 immediately,
2. keeps serving normally for `SHUTDOWN_DELAY` (default 0 locally; about 5 seconds in Kubernetes),
3. then calls `Shutdown`, which stops accepting connections and waits up to `SHUTDOWN_TIMEOUT` for in-flight requests.

Module 15 sets these values in the Deployment manifest, along with probes on `/healthz` and `/readyz`.

## What the tests cover

```bash
go test ./...
```

| Test | Checks |
|---|---|
| `TestCreateResolveRedirect` | the happy path: 201 with `Location`, JSON fields, 302 redirect |
| `TestErrorMapping` | 409, 422, 400, 415, 404 and 405, table-driven |
| `TestServerErrorsAreNotLeaked` | internal errors are logged, not shown; problem includes the request ID |
| `TestRateLimit` | burst of 3 then 429; health checks never limited |
| `TestReadiness` | `/readyz` is 503 until ready |
| `TestMetricsAndLogs` | counters by route pattern and status; access log fields |
| `TestRun` | the real program over real HTTP, and a clean shutdown on cancel |
| `config` tests | defaults, and every invalid variable reported at once |

## Your turn

Ideas to extend the milestone, in rough order of difficulty:

1. Reserve `api`, `healthz` and `readyz` so they cannot be used as custom codes. Where does that rule belong: `link` or `httpapi`?
2. Design `DELETE /api/links/{code}`: which status codes, and which new `Store` method? Then compare with the version milestone 4 adds (lesson 15.4 needs it).
3. Count redirects per code. Should that be a Prometheus label? (Lesson 12.5 says no. Why?) Where should it be stored instead?
4. Add OpenTelemetry tracing with the middleware from lesson 12.5, or with `otelhttp`.

Next, Module 13 replaces the in-memory store with PostgreSQL, without changing a line of `httpapi`.
