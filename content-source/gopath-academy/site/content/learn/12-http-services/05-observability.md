---
title: "12.5 Observability: logs, metrics and traces"
linkTitle: "12.5 Observability"
weight: 5
description: "The three signals of observability for a Go service: structured logs with slog, RED metrics with the Prometheus client, and distributed traces with OpenTelemetry, plus health checks and runtime metrics."
goVersion: "1.27"
---

## What you will learn

- The three signals, **logs**, **metrics** and **traces**, and what each is for
- Structured logs that machines can query
- RED metrics with the Prometheus client, and the cardinality trap
- Distributed tracing with OpenTelemetry: spans, context propagation, `traceparent`
- Liveness and readiness endpoints, and runtime metrics

## The three signals

When a service misbehaves in production, you cannot attach a debugger. You can only look at what it emits:

| Signal | Answers | Example |
|---|---|---|
| **Metrics** | Is something wrong? How much, since when? | 2% of `POST /api/links` returned 500 in the last 5 minutes; p99 latency doubled at 14:02 |
| **Logs** | What exactly happened in this case? | `request failed error="connection refused" request_id=a1b2...` |
| **Traces** | Where did the time go across services? | the request spent 1.8 s of its 2 s waiting on the database |

Metrics are cheap aggregates you alert on. Logs are detailed records of events. Traces follow one request through several services. A request ID, or better a **trace ID** present in all three, connects them.

## Logs: structured, one event per line

Lesson 8.4 introduced `log/slog`, and lesson 12.2 built the access log. For production:

- Use the **JSON handler**. Log platforms (Loki, Elasticsearch, CloudWatch) index the fields, so you can query `status >= 500 AND route = "POST /api/links"`.
- Write to **standard output**. In containers, the platform collects stdout; the twelve-factor app never manages log files.
- Log **events**, not narratives: one line per request, one per failure, with fields rather than formatted sentences.
- Choose levels deliberately: `ERROR` for failures someone should look at, `INFO` for normal events like requests and start-up, `DEBUG` for details you turn on temporarily with `LOG_LEVEL=debug`.

## Metrics with Prometheus

**Prometheus** is the standard metrics system in the Kubernetes world. Your service exposes current metric values as text at `/metrics`; Prometheus **scrapes** that endpoint every few seconds and stores the time series.

For a request-driven service, start with the **RED** metrics: **R**ate (requests per second), **E**rrors (failed requests per second), **D**uration (latency distribution). Three metrics cover all of it:

{{< example file="examples/thirdparty/observability/metrics.go" region="metrics" >}}

This example uses the official client library `github.com/prometheus/client_golang`, so it lives in the `examples/thirdparty` module (the main course module has no dependencies). Run `go mod tidy` in `examples/thirdparty` once to download the dependencies. The metric types:

- **Counter**: only goes up (`http_requests_total`). Prometheus computes rates from it: `rate(http_requests_total[5m])`.
- **Gauge**: goes up and down (`http_requests_in_flight`).
- **Histogram**: counts observations in buckets (`http_request_duration_seconds`), from which Prometheus estimates percentiles: `histogram_quantile(0.99, ...)`.

The test scrapes the registry like Prometheus would:

{{< example file="examples/thirdparty/observability/observability_test.go" region="metricstest" run="cd examples/thirdparty && go test -run TestMetrics -v ./observability" >}}

```text
# TYPE http_requests_total counter
http_requests_total{code="200",route="GET /items/{id}"} 3
http_requests_total{code="404",route="unmatched"} 1
```

### The cardinality trap

Every distinct combination of label values is a separate time series that Prometheus stores. Labelling by the **route pattern** gives a handful of series. Labelling by the **raw path** (`/items/1`, `/items/2`, ...) creates a new series for every ID ever requested, and eventually takes down your monitoring. Never put unbounded values (IDs, user names, raw paths, error messages) in labels. Here `r.Pattern` gives the route for free, and unmatched paths share one label value.

## Traces with OpenTelemetry

A **trace** is the tree of operations performed for one request, possibly across many services. Each operation is a **span**: a name, start and end time, attributes, and a parent. **OpenTelemetry** (OTel) is the vendor-neutral standard for producing traces (and metrics and logs), supported by Jaeger, Grafana Tempo, Honeycomb, Datadog and the cloud providers.

For a trace to cross service boundaries, the caller sends its trace context in the W3C **`traceparent`** header: `00-<trace id>-<parent span id>-<flags>`. The receiving service **extracts** it and creates its spans as children:

{{< example file="examples/thirdparty/observability/tracing.go" region="tracing" >}}

The test sends a request with a `traceparent`, as an upstream service would, and inspects the recorded spans:

{{< example file="examples/thirdparty/observability/observability_test.go" region="tracetest" run="cd examples/thirdparty && go test -run TestTracing -v ./observability" >}}

```text
server span "GET /items/{id}" trace=4bf92f3577b34da6a3ce929d0e0e4736 parent=00f067aa0ba902b7 attrs=[{"http.route":"GET /items/{id}"} {"http.request.method":"GET"} {"http.response.status_code":200}]
child span "load item" parent=c251016ee117c350
```

The server span joined the caller's trace (same trace ID, the caller's span as parent), and the span the handler started from `r.Context()` became its child. That is the whole mechanism: **pass the context**, and spans nest correctly.

In real services:

- The **OTel SDK** is configured once in `main` with an **exporter** that sends spans to a collector (usually OTLP over gRPC or HTTP), and a **sampler** so you do not record every request at high traffic.
- You rarely write the HTTP middleware yourself. The contrib package `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` provides server middleware and a client transport that injects `traceparent` into outgoing requests. Most database drivers and RPC libraries have OTel instrumentation too. The middleware in this lesson exists to show what those packages do.
- Use the attribute names from the OTel **semantic conventions** (`http.route`, `http.response.status_code`, as above), available as constants in `go.opentelemetry.io/otel/semconv`, so dashboards work across services.
- Put the trace ID into your log lines, and logs and traces link up.

## Health checks

Orchestrators need to ask two questions, which deserve separate endpoints:

- **Liveness** (`/healthz`): is the process alive, or stuck and in need of a restart? Keep it trivial. Do **not** check dependencies here: if the database is down, restarting every instance of your service makes things worse.
- **Readiness** (`/readyz`): should this instance receive traffic right now? Report not ready during start-up and **during shutdown** (lesson 12.8), and optionally when a critical dependency is unreachable.

Module 15 wires these into Kubernetes probes.

## Runtime metrics

The Prometheus client's Go collector exports the runtime's own metrics (from `runtime/metrics`, lesson 11.1): goroutine count, GC cycles and pause times, heap size, GOMAXPROCS. Register it and you can graph exactly the things Module 11 explained:

```go
reg := prometheus.NewRegistry()
reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
```

A steadily growing `go_goroutines` is a goroutine leak (lesson 7.6); a growing heap after GC is a memory leak; frequent GC cycles point to allocation churn (lesson 10.4).

## Where to expose it

Metrics, pprof (lesson 10.2) and similar endpoints reveal internals. Serve them on a **separate port** (an "admin" port) that is reachable by Prometheus inside the cluster but never routed publicly. The shortener listens on `:8080` for the API and `:9090` for admin.

## Common mistakes

### Unbounded label values

One label with user IDs can create millions of series. Use route patterns and status codes.

### Logging instead of counting

Counting errors by grepping logs is slow and expensive. Emit a metric and alert on it; use logs for the details.

### Liveness checks that test dependencies

A database outage then becomes a restart storm.

### Breaking the context chain

A handler that starts work with `context.Background()` instead of `r.Context()` starts a new, disconnected trace, and loses cancellation too.

## Key takeaways

- Metrics tell you something is wrong, logs tell you what happened, traces tell you where time went.
- Log structured JSON to stdout, one event per line, with a request or trace ID.
- Expose RED metrics with the Prometheus client, labelled by route pattern and status, never by raw path.
- OpenTelemetry traces cross services via the `traceparent` header; pass `r.Context()` and spans nest.
- Separate liveness from readiness, and serve metrics and pprof on an internal admin port.

## Further reading

- [Prometheus: instrumenting a Go application](https://prometheus.io/docs/guides/go-application/)
- [OpenTelemetry Go documentation](https://opentelemetry.io/docs/languages/go/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
