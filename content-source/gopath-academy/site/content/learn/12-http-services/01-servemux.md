---
title: "12.1 net/http and ServeMux routing"
linkTitle: "12.1 ServeMux routing"
weight: 1
description: "How a Go HTTP server is put together (Handler, ServeMux, Server), and the routing patterns of Go 1.22's ServeMux: methods, wildcards, exact matches, precedence and conflicts."
goVersion: "1.27"
---

## What you will learn

- The three building blocks of a Go HTTP server: `Handler`, `ServeMux` and `Server`
- Routing patterns: methods, `{wildcards}`, `{rest...}`, `{$}` and subtrees
- How ServeMux chooses between overlapping patterns, and when it refuses
- Automatic `405 Method Not Allowed` and `HEAD` handling
- How to test handlers without opening a port

## The building blocks

Everything in `net/http` server code is built on one interface:

```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

A handler receives the request and writes the response. `http.HandlerFunc` turns an ordinary function with that signature into a `Handler`, which is why most code just writes functions.

A **`ServeMux`** is a handler that dispatches to other handlers based on the request's method, host and path: a router. An **`http.Server`** listens on a port, parses HTTP, and calls one handler (usually a ServeMux) for each request, **each in its own goroutine**. That last point matters: handlers run concurrently, so any state they share needs synchronization (Module 6).

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello")
})
srv := &http.Server{Addr: ":8080", Handler: mux} // timeouts: lesson 12.6
log.Fatal(srv.ListenAndServe())
```

You will see `http.ListenAndServe(":8080", mux)` in tutorials. It creates a `Server` with **no timeouts**, which is fine for experiments but not for anything exposed to a network (lesson 12.6).

## Routing patterns (Go 1.22)

Before Go 1.22, ServeMux matched only path prefixes, and every project reached for a third party router to get methods and path parameters. Go 1.22 added both. A pattern has the form:

```text
[METHOD ][HOST]/[PATH]
```

This mux uses every kind of pattern:

{{< example file="examples/m12/routing/routing.go" region="mux" >}}

And this is what it does with a set of requests:

{{< example file="examples/m12/routing/routing_test.go" region="example" run="go test -run Example_routing -v ./examples/m12/routing" >}}

Reading the results line by line:

- **`GET /{$}`** matches only `/`. A pattern ending in a slash matches a whole subtree, so a plain `"/"` would match every path. That surprises everyone once; `{$}` means "end of path".
- **`/nope`** matches nothing: 404, with ServeMux's default body.
- **`{id}`** matches one path segment. `r.PathValue("id")` returns it, already unescaped.
- **`/items/new`** goes to its own handler even though `/items/{id}` also matches. See precedence below.
- **`DELETE /items`**: the path exists, but only for `POST`. ServeMux answers `405 Method Not Allowed` with an `Allow` header listing the methods that would work. You get this for free by putting methods in your patterns.
- **`{path...}`** at the end matches all remaining segments, slashes included: `docs/a/b.txt`.
- **`/static/`**, with no method, matches any method and everything under `/static/`.

Two more details: a `GET` pattern also matches `HEAD` requests (the server then discards the body), and a pattern can start with a host, `"api.example.com/"`, to route by the `Host` header.

## Precedence and conflicts

When several patterns match a request, the **most specific** one wins. A pattern is more specific than another if it matches a strict subset of its requests. `/items/new` matches only one path, `/items/{id}` matches many, so `/items/new` wins. Registration order never matters.

But two patterns can overlap without either being more specific:

{{< example file="examples/m12/routing/routing_test.go" region="conflict" >}}

`/posts/{id}` and `/{resource}/latest` both match `/posts/latest`, and neither is a subset of the other. Rather than guess, ServeMux **panics at registration** with a clear explanation:

```text
panic: pattern "GET /{resource}/latest" (registered at .../main.go:8) conflicts with pattern "GET /posts/{id}" (registered at .../main.go:7):
	GET /{resource}/latest and GET /posts/{id} both match some paths, like "/posts/latest".
	But neither is more specific than the other.
```

Panicking at start-up is a feature: an ambiguous route table fails the first test run, not a request in production.

## Testing handlers

`net/http/httptest` lets you call a handler directly, with no network:

```go
rec := httptest.NewRecorder()                                      // a ResponseWriter that records
mux.ServeHTTP(rec, httptest.NewRequest("GET", "/items/42", nil))   // a ready-made *Request
fmt.Println(rec.Code, rec.Body.String())                           // 200 item 42
```

Every example in this module is tested this way. For tests that need a real network connection (timeouts, TLS, a client under test), `httptest.NewServer` starts a server on a free local port; lesson 12.6 uses one.

## Do you need a framework?

For most JSON APIs, no. ServeMux now covers methods, path parameters and precedence; middleware is a plain function (lesson 12.2); JSON is `encoding/json` (lesson 12.3). Routers such as chi build on the same `http.Handler` interface, so moving between them is cheap. Frameworks that use their own handler and context types (Gin, Echo, Fiber) offer conveniences, but tie your code to their APIs. Start with the standard library and add a dependency when you can name what it gives you.

## Common mistakes

### Registering "/" for the home page

It matches every unknown path, which then returns your home page with 200 instead of 404. Use `"GET /{$}"`.

### Using http.ListenAndServe in production

It has no timeouts. Build an `http.Server` (lesson 12.6).

### Sharing state between handlers without locking

Each request runs in its own goroutine. Run your tests with `-race`.

### Using the DefaultServeMux

`http.HandleFunc` without a mux registers on a global ServeMux that any imported package can also register on (`net/http/pprof` does). Create your own with `http.NewServeMux()`.

## Key takeaways

- A `Handler` serves one request; a `ServeMux` routes to handlers; an `http.Server` runs each request in its own goroutine.
- Patterns: `"METHOD /path/{name}"`, `{rest...}` for the remaining path, `{$}` for an exact end, a trailing `/` for a subtree; read wildcards with `r.PathValue`.
- The most specific pattern wins regardless of order; ambiguous patterns panic at registration.
- Methods in patterns give you `405` with an `Allow` header automatically.
- Test handlers with `httptest.NewRecorder` and `httptest.NewRequest`.

## Further reading

- [Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements) on the Go blog
- [Package net/http: ServeMux](https://pkg.go.dev/net/http#ServeMux), with the full pattern syntax and precedence rules
