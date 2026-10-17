---
title: "12. Building HTTP services"
weight: 12
description: "Production HTTP services with the standard library: ServeMux routing, middleware, JSON APIs with proper validation and errors, configuration, logs, metrics and traces, and the security settings every server needs."
---

Part 4 is about running Go in production. It starts where most Go services start: an HTTP API. Go's standard library is enough to build one properly. Since Go 1.22 its router understands methods and path parameters, so many teams no longer need a web framework at all.

This module builds the pieces one at a time, each with tests, and then assembles them into **milestone 2 of the URL shortener**: a complete service with JSON endpoints, redirects, request IDs, access logs, Prometheus metrics, rate limiting, health checks and graceful shutdown. Module 15 deploys it to Kubernetes.

**Time:** about 5 to 6 hours, plus the exercise and project.

{{< cards >}}
  {{< card link="01-servemux" title="12.1 net/http and ServeMux routing" subtitle="Methods, wildcards, precedence." >}}
  {{< card link="02-middleware" title="12.2 Middleware" subtitle="Request IDs, access logs, panic recovery." >}}
  {{< card link="03-json-apis" title="12.3 JSON APIs and validation" subtitle="Strict decoding, problem details, status codes." >}}
  {{< card link="04-configuration" title="12.4 Configuration" subtitle="Twelve-factor config and secrets." >}}
  {{< card link="05-observability" title="12.5 Observability" subtitle="Logs, Prometheus metrics, OpenTelemetry traces." >}}
  {{< card link="06-security" title="12.6 Security basics" subtitle="Timeouts, limits, TLS, CSRF, headers." >}}
  {{< card link="07-exercise" title="12.7 Exercise: a notes API" subtitle="Build a JSON CRUD API from a spec." >}}
  {{< card link="08-project-shortener-api" title="12.8 Project: shortener HTTP API" subtitle="Everything together, tested end to end." >}}
{{< /cards >}}
