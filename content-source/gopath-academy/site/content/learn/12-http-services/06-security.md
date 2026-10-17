---
title: "12.6 Security basics: timeouts, limits, TLS and CSRF"
linkTitle: "12.6 Security basics"
weight: 6
description: "The settings every Go HTTP server exposed to a network needs: server timeouts (with a slowloris test), header and body limits, TLS configuration, security headers, Go 1.25's CrossOriginProtection against CSRF, and handling secrets."
goVersion: "1.27"
---

## What you will learn

- Why `http.Server`'s zero values are unsafe on the internet, and which timeouts to set
- How a slowloris attack works, and a test that shows the defense
- Limiting header and body sizes
- TLS in Go, and when to terminate it elsewhere
- Security headers, and CSRF protection with `http.CrossOriginProtection` (Go 1.25)

## Timeouts

`http.Server` with no timeouts waits **forever** for a client. One client can open thousands of connections and send nothing, or send a byte every few seconds, and the server keeps a goroutine and a socket for each until it runs out of file descriptors or memory. Set all four timeouts:

{{< example file="examples/m12/security/security.go" region="server" >}}

| Field | Limits |
|---|---|
| `ReadHeaderTimeout` | reading the request line and headers. The most important one. |
| `ReadTimeout` | reading the entire request, body included. |
| `WriteTimeout` | from the end of reading the headers to the end of writing the response. It caps how long a handler can take, so streaming endpoints need a different value (set per request with `http.ResponseController.SetWriteDeadline`). |
| `IdleTimeout` | how long a keep-alive connection may wait for its next request. |

### Slowloris, tested

A **slowloris** client sends headers slowly and never finishes, to tie up the server. This test does exactly that against a real server:

{{< example file="examples/m12/security/security_test.go" region="slowloris" run="go test -v -run TestReadHeaderTimeout ./examples/m12/security" >}}

```text
security_test.go:37: server gave up after 200ms (err=EOF)
```

With `ReadHeaderTimeout` set, the server closes the connection when the time runs out. Remove that line and the test fails after 5 seconds: the server would have waited forever.

Timeouts are also needed in the other direction. Every outgoing `http.Client` needs a `Timeout` and every outgoing request a context with a deadline (lesson 8.7), or one slow dependency can make all your handlers hang.

## Size limits

- **Headers**: `MaxHeaderBytes` defaults to 1 MiB. 64 KiB is plenty for most APIs.
- **Bodies** have no default limit. Limit them per handler with `http.MaxBytesReader` (lesson 12.3) or for a whole server with `http.MaxBytesHandler`:

{{< example file="examples/m12/security/security.go" region="limit" >}}

Also bound everything a client controls in size or count: list page sizes, batch sizes, upload sizes, query complexity.

## TLS

Go's `crypto/tls` is a complete, well-maintained TLS implementation with secure defaults: TLS 1.2 as the minimum for servers, modern cipher suites only, TLS 1.3 preferred, and post-quantum key exchange enabled by default in recent releases. To serve HTTPS directly:

```go
srv := NewServer(":8443", handler)
log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))
```

You rarely need to change the `tls.Config`; setting `MinVersion` explicitly, as above, documents the intent. For automatic certificates from Let's Encrypt, the `golang.org/x/crypto/acme/autocert` package handles issuance and renewal.

In **Kubernetes**, TLS is usually terminated at the **ingress controller** or a service mesh, and the Go service speaks plain HTTP inside the cluster (Module 15). That is a reasonable default; teams that need encryption inside the cluster add mesh mTLS rather than certificates in every service.

## Security headers

APIs consumed by browsers should send a few headers that make browsers stricter:

{{< example file="examples/m12/security/security.go" region="headers" >}}

`nosniff` stops browsers from interpreting a JSON response as HTML or script; the CSP tells them this response may load nothing and may not be framed; HSTS tells them to use HTTPS for this host from now on, and is only sent over TLS.

## CSRF protection (Go 1.25)

**Cross-site request forgery**: a malicious page makes the victim's browser send a request to your site, and the browser attaches the victim's cookies. If your site uses cookie sessions, the request is authenticated. The traditional defense was CSRF tokens in every form.

Go 1.25 added `http.CrossOriginProtection`, which uses headers that all browsers since 2023 send: **`Sec-Fetch-Site`**, falling back to comparing **`Origin`** with `Host`. It rejects state-changing requests (anything except GET, HEAD and OPTIONS) that come from another origin:

{{< example file="examples/m12/security/csrf.go" region="csrf" >}}

{{< example file="examples/m12/security/csrf_test.go" region="csrftest" run="go test -run ExampleProtectForms -v ./examples/m12/security" >}}

Same-origin POSTs pass, cross-site POSTs get 403, safe methods always pass, trusted origins pass, and requests without browser headers (curl, other servers) pass because they cannot carry a victim's cookies. Two consequences:

- Your GET handlers must **never change state**, because they are always allowed.
- APIs authenticated with an `Authorization` header instead of cookies are not exposed to CSRF in the first place, since browsers do not attach that header automatically.

These files carry `//go:build go1.25`, because the course's main module still supports Go 1.24.

## Secrets

- Read secrets from files or the environment at start-up (lesson 12.4); never log them, never put them in URLs (URLs end up in logs and browser history).
- Compare secrets such as API keys with `crypto/subtle.ConstantTimeCompare`, so the comparison time does not leak how many characters matched.
- Generate tokens and IDs with `crypto/rand` (`rand.Text()` since Go 1.24 returns a secure random string), never `math/rand`.

## Common mistakes

### No timeouts

`http.ListenAndServe` and a bare `&http.Server{}` have none. Always set them.

### Unlimited bodies

`io.ReadAll(r.Body)` on a public endpoint reads whatever a client sends.

### State changes on GET

Browsers, crawlers and link previewers follow GET links freely, and CSRF protection lets them through.

### Comparing secrets with ==

Use `subtle.ConstantTimeCompare`.

## Key takeaways

- Set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout` and `IdleTimeout` on every server; the zero values mean "wait forever".
- Limit header size with `MaxHeaderBytes` and body size with `MaxBytesReader` or `MaxBytesHandler`.
- Go's TLS defaults are good; in Kubernetes, TLS usually terminates at the ingress.
- Send `nosniff`, a restrictive CSP and, over TLS, HSTS.
- Wrap cookie-authenticated, state-changing endpoints with `http.CrossOriginProtection` (Go 1.25), and keep GET handlers free of side effects.

## Further reading

- [The complete guide to Go net/http timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) on the Cloudflare blog
- [Package net/http: CrossOriginProtection](https://pkg.go.dev/net/http#CrossOriginProtection)
- [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/)
