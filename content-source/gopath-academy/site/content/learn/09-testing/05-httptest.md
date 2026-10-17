---
title: "9.5 httptest"
linkTitle: "9.5 httptest"
weight: 5
description: "Test HTTP handlers directly with httptest.NewRecorder, test clients and full stacks with httptest.NewServer, and use the in-memory NewTestServer from Go 1.27."
goVersion: "1.27"
---

## What you will learn

- Unit testing handlers with `httptest.NewRequest` and `httptest.NewRecorder`
- End-to-end tests with `httptest.NewServer`
- Which style to use when
- Testing clients against fake servers
- `httptest.NewTestServer` (Go 1.27) for in-memory, synctest-friendly servers

## Handlers under test

{{< example file="examples/m09/httptesting/handler.go" region="handler" >}}

## Unit tests with `ResponseRecorder`

{{< example file="examples/m09/httptesting/handler_test.go" region="recorder" run="go test -v -run TestGreetHandler$ ./examples/m09/httptesting" >}}

- `httptest.NewRequest(method, target, body)` builds an incoming server request. Unlike `http.NewRequest`, it panics on errors (fine in tests) and fills in fields a server would set.
- `httptest.NewRecorder()` is an `http.ResponseWriter` that records the status code (`rec.Code`), headers (`rec.Header()`) and body (`rec.Body`).
- You call the handler directly: no network, no goroutines, microseconds per case. This is the right default for handler logic.

## End-to-end tests with `NewServer`

{{< example file="examples/m09/httptesting/handler_test.go" region="server" run="go test -v -run TestGreetOverHTTP ./examples/m09/httptesting" >}}

`httptest.NewServer(handler)` starts a real HTTP server on a random localhost port. Use it to test:

- **routing, middleware and the full handler stack** (`ServeMux`, authentication, logging);
- **your HTTP clients** against a fake API (lesson 8.7 and its exercise);
- behavior that depends on real connections: timeouts, streaming, redirects, TLS (`httptest.NewTLSServer`).

`srv.Client()` returns a client configured for the server (it trusts the test TLS certificate). Always `defer srv.Close()` or register it with `t.Cleanup`.

## Which one?

| Test | Use |
|---|---|
| A handler's own logic: validation, status codes, body | `NewRecorder` |
| Routing, middleware chain, path parameters | `NewServer` (or call the mux with a recorder) |
| An HTTP client | `NewServer` with a fake handler |
| Timeouts and cancellation | `NewServer`, or `NewTestServer` with `synctest` |

## `httptest.NewTestServer` (Go 1.27)

Go 1.27 added `httptest.NewTestServer(t, handler)`, which runs the server over an **in-memory network** instead of a real socket. Combined with `testing/synctest` (lesson 9.9), timeouts in clients and servers become deterministic and instant: a test of a 30-second timeout runs on the virtual clock and finishes immediately. It also registers its own cleanup with `t`.

## Common mistakes

### Forgetting to check the status code

A handler test that only checks the body passes even when the status is wrong. Check both.

### Comparing JSON as strings

Encoders may change spacing or key order between versions. Decode the body and compare values, or compare against a canonical form.

### Not closing the test server

Leaks a listener and goroutines per test. Use `t.Cleanup(srv.Close)`.

## Exercise

{{< exercise dir="m09/fizzhttp" >}}
Implement an HTTP handler for `GET /fizzbuzz?n=N` that validates input, sets headers, returns JSON and rejects other methods with `405` and an `Allow` header. The tests use both `NewRecorder` and `NewServer`.
{{< /exercise >}}

## Key takeaways

- Test handler logic directly with `NewRequest` and `NewRecorder`.
- Use `NewServer` for routing, middleware, real connections and client tests.
- Check status, headers and body; decode JSON instead of comparing strings.
- Go 1.27's `NewTestServer` uses an in-memory network that works with `synctest`.

## Further reading

- [Package net/http/httptest](https://pkg.go.dev/net/http/httptest)
- [Go 1.27 release notes](https://go.dev/doc/go1.27)
