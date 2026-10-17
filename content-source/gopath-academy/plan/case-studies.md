# Case studies and projects

## Rules for company case studies

1. Every factual claim links to a primary source: the company's engineering blog, a conference talk, a paper, or the official Go case study page (go.dev/solutions).
2. No invented numbers. If a source does not give a number, we do not give one.
3. Each case study ends with "What you can apply" tying the story back to specific lessons.
4. Paraphrase sources. Short quotes only, always attributed.

## Company case studies (candidates)

Status: `[ ]` not researched, `[~]` sources collected, `[x]` written and reviewed.

| Status | Page | Angle | Lessons it connects to |
|---|---|---|---|
| [x] | `kubernetes-and-docker` | Why the defining cloud native tools were written in Go; the 2013 pain points and what fixed them | 1.1, 1.4, 6.x, 14.1, 15.x |
| [x] | `dl-google-com` | C++ to Go rewrite: incremental, audit mode, `ServeContent`, less than half the code | 4.x, 8.1, 12.x, 13.5 |
| [x] | `cloudflare` | Early production Go: goroutine per request, recovered panics | 5.5, 6.4, 12.2 |
| [x] | `dropbox` | Performance-critical backends from Python to Go; the library gap | 5.x, 10.x, 13.x |
| [x] | `uber-geofence` | 170k QPS with a two-level index and an RWMutex | 6.4, 6.5, 7.x, 10.x, 12.5 |
| [x] | `uber-data-races` | PLDI 2022: 2,000 races in 46M lines, the Go-specific patterns, what Go 1.22 and 1.25 fixed | 2.5, 6.4, 6.6 |
| [x] | `twitch` | Memory ballast, GC assists, and GOMEMLIMIT as the modern answer | 10.4, 11.3 |
| [x] | `discord` | Go to Rust for one service: forced GC every 2 minutes over a huge cache; a fair reading | 1.1, 11.3 |
| [x] | `monzo` | A bank as Go microservices on Kubernetes; idempotent retries, at-least-once delivery | 12.x, 13.8, 15.x |
| [x] | go.dev/solutions | Linked from the case studies index for further reading | |

All sources were re-checked against the original pages while writing (October 2026). Uber was split into two studies because the geofence post and the data race paper teach different things.

Ideas for later: PayPal, American Express or Riot Games from go.dev/solutions; Grafana Labs or HashiCorp as open source Go companies; Go at Google (the 2012 "Go at Google" paper).

Each candidate needs its sources verified before writing. Some of these may not hold up on research; we drop those rather than stretch them.

## Build-along projects

Each project is built in milestones. Every milestone has its own folder under `projects/` with code that compiles and passes tests.

| Status | Project | Concepts practiced | After module |
|---|---|---|---|
| [x] | `greet` CLI | Modules, packages, flags, ldflags, cross-compiling, GoReleaser | 1 and 14 |
| [x] | Word frequency counter (`wordfreq`) | Strings, maps, sorting, `io.Reader` | 3 |
| [x] | URL shortener (`shortener`, 4 milestones) | Domain errors, `net/http`, PostgreSQL, Redis, Kubernetes | 5, 12, 13, 15 |
| [x] | Concurrent web crawler (`crawler`) | Goroutines, channels, `context`, rate limiting | 7 |
| [x] | Token bucket rate limiter library (`ratelimit`) | Concurrency primitives, benchmarks | 7 and 10 |
| [x] | Mini Redis (`miniredis`) | TCP servers, parsing, concurrency, persistence | 8 |
| [x] | Background job queue (`jobqueue`) | Worker pools, retries, `FOR UPDATE SKIP LOCKED`, graceful shutdown | 13 |
| [x] | `shortctl` CLI | Cobra, config files, the shortener's API | 14 |
| [x] | ShortLink operator (`shortlink-operator`) | CRDs, controller-runtime, finalizers, leader election | 15 |
