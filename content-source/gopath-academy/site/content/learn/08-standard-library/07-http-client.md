---
title: "8.7 net/http client done right"
linkTitle: "8.7 HTTP client"
weight: 7
description: "Make HTTP requests that cannot hang, leak connections or silently ignore errors: timeouts, contexts, closing and limiting bodies, checking status codes, and tuning the Transport."
goVersion: "1.27"
---

## What you will learn

- Why the default HTTP client is a production risk
- The steps every correct request needs
- Timeouts: `Client.Timeout`, request contexts and `Transport` timeouts
- Why you must close response bodies, and why status codes are not errors
- Testing clients against `httptest.Server`
- The `Transport` settings worth knowing

## The idea

`net/http` makes requests look trivial: `resp, err := http.Get(url)`. That line has three production problems:

1. The default client has **no timeout**. A server that accepts the connection and never answers blocks your goroutine forever.
2. If you forget `resp.Body.Close()`, the connection cannot be reused, and you leak file descriptors and goroutines.
3. A `404` or `500` response returns `err == nil`. If you only check `err`, failures look like successes.

## A correct request, step by step

{{< example file="examples/m08/httpclient/httpclient_test.go" region="client" >}}

1. **One shared `http.Client`** for the whole program. Clients are safe for concurrent use and keep a pool of connections. Creating a client per request throws that pool away.
2. **`http.NewRequestWithContext`** so the caller's cancellation and deadline apply (lesson 7.1).
3. **Set headers** such as `Accept`, authentication and a `User-Agent`.
4. **Check `err` from `Do`.** It covers network failures, timeouts and cancellation, never HTTP status codes.
5. **`defer resp.Body.Close()`** immediately after the error check.
6. **Check `resp.StatusCode`** and turn unexpected statuses into errors.
7. **Limit and decode the body.** `io.LimitReader` stops a misbehaving server from filling your memory.

For connection reuse, the body should be read to the end before closing; since Go 1.27, closing an HTTP/1 response body drains a limited amount of unread content automatically, but reading or discarding what you do not need (as the error path does with `io.Copy(io.Discard, ...)`) keeps the behavior explicit.

## Testing against a fake server

{{< example file="examples/m08/httpclient/httpclient_test.go" region="test" run="go test -run Example_request -v ./examples/m08/httpclient" >}}

`httptest.NewServer` starts a real HTTP server on a local port with any handler you give it. Tests run fast, need no network, and can simulate every failure: slow responses, bad JSON, 500s. Go 1.27 adds `httptest.NewTestServer`, which uses an in-memory network and works with `testing/synctest` for deterministic timing.

## Timeouts

{{< example file="examples/m08/httpclient/httpclient_test.go" region="timeout" run="go test -run Example_timeouts -v ./examples/m08/httpclient" >}}

There are several layers, and they combine:

| Setting | Limits | Use |
|---|---|---|
| `Client.Timeout` | the entire request, including reading the body | a safety net on every client |
| Request context deadline | the entire request, per call | when each call needs its own budget |
| `Transport.DialContext` timeout | establishing the TCP connection | detecting unreachable hosts quickly |
| `Transport.TLSHandshakeTimeout` | the TLS handshake | slow or broken TLS endpoints |
| `Transport.ResponseHeaderTimeout` | waiting for response headers after sending the request | servers that accept but never answer |

A timeout error satisfies `net.Error` with `Timeout() == true`, and a context deadline also matches `context.DeadlineExceeded`. Both are worth retrying for idempotent requests, with backoff (lesson 5.6).

## Tuning the Transport

{{< example file="examples/m08/httpclient/httpclient_test.go" region="transport" run="go test -run Example_transport -v ./examples/m08/httpclient" >}}

Clone `http.DefaultTransport` and adjust it instead of building one from scratch, so you keep its sensible defaults (proxy from environment, HTTP/2, keep-alives). The setting people most often need is **`MaxIdleConnsPerHost`**: its default of 2 means a service making many concurrent calls to one backend keeps opening new connections. Raise it to roughly your expected concurrency to that host.

## Retries

Only retry requests that are **safe to repeat**: `GET`, `HEAD`, `PUT` and `DELETE` are idempotent by definition; `POST` usually is not, unless the API supports idempotency keys. Retry on network errors, timeouts, `429` and `5xx`, with exponential backoff and jitter, and respect a `Retry-After` header.

## Common mistakes

### Using `http.Get` or `http.DefaultClient` in services

No timeout. Use your own client with `Timeout` set.

### A new `http.Client` (or `Transport`) per request

Destroys connection reuse. Every request pays for a TCP and TLS handshake, and idle connections pile up.

### Not closing the body

Leaks connections. `defer resp.Body.Close()` right after checking `err`.

### Treating non-2xx as success

`err == nil` with a 503 response. Always check `StatusCode`.

### Reading unlimited bodies

`io.ReadAll(resp.Body)` from a server you do not control. Wrap it in `io.LimitReader`.

## Exercise

{{< exercise dir="m08/apiclient" >}}
Write a small API client with `GetUser` and `CreateUser`. It must send the right headers, respect the context, map `404` to `ErrNotFound` and other failures to an `*APIError` that carries the server's error message, and decode JSON responses. The tests run against an `httptest` server, including a slow endpoint that must time out.
{{< /exercise >}}

## Key takeaways

- Use one shared `http.Client` with a `Timeout`; never the default client in services.
- Every request: context, headers, check `err`, close the body, check the status, limit the body.
- Status codes are not errors; translate them yourself.
- Test against `httptest.Server`; tune `MaxIdleConnsPerHost` for busy backends.
- Retry only idempotent requests, with backoff.

## Further reading

- [Package net/http](https://pkg.go.dev/net/http), especially Client and Transport
- [Package net/http/httptest](https://pkg.go.dev/net/http/httptest)
- [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) on the Cloudflare blog
