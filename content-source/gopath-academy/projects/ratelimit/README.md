# ratelimit

Token bucket rate limiter built in Module 7 (lesson 7.5 and 7.8), benchmarked in Module 10.

- `New(rate, burst)`: a bucket with `Allow`, `AllowN`, `Wait`, `WaitN`, `Tokens`
- `NewKeyed(rate, burst)`: one bucket per key, with `Prune` for idle keys

```bash
go test -race ./...
go test -bench . -benchmem
```
