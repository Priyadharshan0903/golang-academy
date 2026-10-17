---
title: "12.2 Middleware"
linkTitle: "12.2 Middleware"
weight: 2
description: "Write HTTP middleware as plain functions: chaining and ordering, capturing the status code safely, request IDs through the context, structured access logs, and panic recovery."
goVersion: "1.27"
---

## What you will learn

- What middleware is in Go: a function from `http.Handler` to `http.Handler`
- How to chain middleware, and why the order matters
- How to capture the response status without breaking `Flush` and friends
- Request IDs carried through `context.Context`
- Access logs with `log/slog`, and turning panics into 500 responses

## The idea

Some behavior applies to every request: logging, metrics, authentication, request IDs, panic recovery. Instead of repeating it in each handler, you **wrap** the handler. In Go, middleware is just this:

{{< example file="examples/m12/middleware/middleware.go" region="chain" >}}

A middleware takes the next handler and returns a new handler that does something before and/or after calling it. No framework, no registration: plain functions.

## Order matters

{{< example file="examples/m12/middleware/middleware_test.go" region="order" run="go test -run ExampleChain -v ./examples/m12/middleware" >}}

Middleware nests like function calls: `Chain(h, A, B)` is `A(B(h))`. The request passes through A, then B, then the handler; the response passes back through B, then A. Decide the order by asking what each middleware needs to see:

- **Request ID** goes outermost, so everything inside can log the ID.
- **Logging** must see the final status code, including the 500 produced when a handler panics, so it goes **outside** panic recovery.
- **Recovery** goes close to the handlers, so it catches their panics and writes a response the outer layers can observe.
- **Authentication** and **rate limiting** go before the expensive work, often only on some routes.

## Capturing the status code

`http.ResponseWriter` does not tell you which status a handler wrote. To log it, wrap the writer:

{{< example file="examples/m12/middleware/middleware.go" region="recorder" >}}

Three details make this wrapper correct:

- If a handler calls `Write` without `WriteHeader`, the status is **200**. The wrapper records that.
- Only the **first** `WriteHeader` counts, like in net/http (a second call is ignored, with a log message).
- **`Unwrap`**. Handlers sometimes need features of the underlying writer: flushing a streaming response, setting per-request deadlines, hijacking for WebSockets. The modern way is `http.NewResponseController(w)`, which calls `Unwrap` to find the original writer through any number of wrappers. Without `Unwrap`, your middleware would silently break streaming. The test `TestResponseControllerThroughRecorder` proves it works.

## Request IDs and the context

A request ID ties together every log line about one request, across services. The middleware reuses a sane incoming `X-Request-ID` (set by a load balancer or calling service) or creates one, and stores it in the request's **context** (lesson 7.1):

{{< example file="examples/m12/middleware/middleware.go" region="requestid" >}}

Notes:

- The context key is an unexported type (`requestIDKey{}`), so no other package can collide with it or read it except through `RequestIDFrom`.
- `r.WithContext` returns a **new** request; the middleware passes that one on.
- Incoming IDs are **validated**: a client controls this header, and writing raw client input to logs invites log injection (fake log lines via newlines) and enormous values.
- Request-scoped values like this are what the context's values are for. Do not use the context to pass optional function parameters.

## Access logs with slog

{{< example file="examples/m12/middleware/middleware.go" region="logging" >}}

One structured line per request, written **after** the handler returns, so it has the status and duration. `r.Pattern` (Go 1.23) is the ServeMux pattern that matched, such as `GET /hello/{name}`: much better for grouping than the raw path, which contains IDs. It is set by the ServeMux on the request it receives, so the logging middleware sees it as long as it passes the same `*Request` down; middleware that calls `r.WithContext` creates a new request, and must sit outside logging, as `RequestID` does here. `LogAttrs` with typed attributes avoids allocations in this hot path (lesson 8.4).

## Recovering from panics

`net/http` already recovers panics in handlers: it logs the stack trace and **closes the connection**, so the client sees a network error rather than a response. A recovery middleware does better: it logs through your structured logger, with the request ID, and sends a proper 500.

{{< example file="examples/m12/middleware/middleware.go" region="recover" >}}

`http.ErrAbortHandler` is a special panic value that handlers use to abort a response deliberately; it is re-panicked so net/http handles it as intended.

Putting it all together, with time and duration removed so the output is stable:

{{< example file="examples/m12/middleware/middleware_test.go" region="logtest" run="go test -run ExampleLogging -v ./examples/m12/middleware" >}}

The panicking request produces an error log line with the panic value (and, in real use, the stack), then an access log line showing status 500. Both carry the request ID.

## Middleware for specific routes

Middleware does not have to wrap the whole mux. Wrap individual handlers when only some routes need it:

```go
mux.Handle("POST /api/links", rateLimit(http.HandlerFunc(createLink)))
mux.Handle("GET /healthz", http.HandlerFunc(healthz)) // never rate limited
```

The shortener (lesson 12.8) does exactly this.

## Common mistakes

### Wrapping the writer without Unwrap

Streaming, `Flush` and deadlines stop working through your middleware. Add `Unwrap() http.ResponseWriter`.

### Logging inside recovery instead of outside

The access log then misses every panicking request. Order: logging outside, recovery inside.

### Trusting client headers in logs

Validate or sanitize request IDs and anything else from the client before logging it.

### Doing work after writing the response in the handler

The client is already waiting for the connection to finish; heavy work after the response delays it and runs without a deadline. Hand it to a background worker (Module 13).

## Key takeaways

- Middleware is `func(http.Handler) http.Handler`; chain it with plain function calls, first is outermost.
- Order by need: request ID outermost, logging outside recovery, route-specific middleware on individual handlers.
- Wrap `ResponseWriter` to capture the status; record 200 for bare writes and implement `Unwrap`.
- Store request-scoped values like request IDs in the context, with an unexported key type.
- Log with `r.Pattern` for the route, and recover panics into logged 500 responses.

## Further reading

- [Package net/http: ResponseController](https://pkg.go.dev/net/http#ResponseController)
- [Package log/slog](https://pkg.go.dev/log/slog)
