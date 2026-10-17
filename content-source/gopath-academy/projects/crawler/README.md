# crawler

Concurrent web crawler, the project for Part 2 (Modules 6 and 7).

- A coordinator goroutine owns the queue and the seen-set; workers only fetch.
- Bounded concurrency, depth and page limits, a shared rate limiter, and `context` cancellation.
- Tested against a local `httptest` site: no network needed.

```bash
go test -race ./...
go run ./cmd/crawler -url http://localhost:1313/ -depth 2   # crawl the course site while running `make serve`
```

Only crawl sites you own or have permission to crawl.
