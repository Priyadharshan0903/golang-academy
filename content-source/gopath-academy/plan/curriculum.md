# Curriculum

Audience: developers who already program in another language and are new to Go, or who have started Go and want depth. Not aimed at absolute programming beginners.

Every lesson follows the template in [lesson-checklist.md](lesson-checklist.md). Tick a box only when the lesson passes that checklist and has been reviewed.

Legend: `[ ]` not started, `[~]` drafted, `[x]` reviewed and done. `(ex)` means the lesson has an exercise.

---

## Part 1: Fundamentals

### Module 1: Setup and tooling

- [~] 1.1 Why Go, and when not to use it
- [~] 1.2 Installing Go and setting up your editor
- [~] 1.3 Your first program, line by line (ex)
- [~] 1.4 Go modules: `go.mod`, `go.sum` and dependencies
- [~] 1.5 Packages, visibility and project layout (ex)
- [~] 1.6 The `go` command toolbox (build, run, install, env, cross-compiling, ldflags)
- [~] 1.7 Formatting, vetting and linting
- [~] 1.8 Checkpoint project: a small CLI with flags and version stamping (ex)

### Module 2: Language basics

- [~] 2.1 Variables, zero values and short declarations
- [~] 2.2 Basic types, numeric conversions and overflow
- [~] 2.3 Constants and `iota`
- [~] 2.4 Strings, bytes, runes and UTF-8 (ex)
- [~] 2.5 Control flow: `if`, `for`, `switch`, labels
- [~] 2.6 Functions: multiple returns, variadics, closures (ex)
- [~] 2.7 `defer`, and how it is evaluated
- [~] 2.8 Pointers without fear (ex)
- [~] 2.9 Scope and shadowing gotchas

### Module 3: Composite types

- [~] 3.1 Arrays and why you rarely use them directly
- [~] 3.2 Slices: length, capacity and the backing array (ex)
- [~] 3.3 Slice gotchas: aliasing, `append` surprises, memory retention
- [~] 3.4 The `slices` and `maps` packages
- [~] 3.5 Maps: usage, iteration order, concurrency warning (ex)
- [~] 3.6 Named types, structs, tags and anonymous structs
- [~] 3.7 Embedding and composition
- [~] 3.8 Iterators and range-over-func (ex)
- [~] 3.9 Checkpoint project: word frequency counter

### Module 4: Methods, interfaces and generics

- [~] 4.1 Methods: value vs pointer receivers
- [~] 4.2 Interfaces: implicit satisfaction and small interfaces (ex)
- [~] 4.3 The empty interface, `any`, type assertions and type switches
- [~] 4.4 Nil interfaces vs interfaces holding nil (classic gotcha)
- [~] 4.5 Common standard interfaces: `Stringer`, `io.Reader`, `io.Writer`, `sort.Interface`
- [~] 4.6 Generics: type parameters and constraints (ex)
- [~] 4.7 Generic methods (Go 1.27) and when generics are the wrong tool
- [~] 4.8 Designing APIs: accept interfaces, return structs

### Module 5: Errors

- [~] 5.1 Errors are values
- [~] 5.2 Wrapping with `%w`, `errors.Is`, `errors.As` (ex)
- [~] 5.3 Sentinel errors vs custom error types
- [~] 5.4 `errors.Join` and multi-errors
- [~] 5.5 `panic` and `recover`: when, and mostly when not
- [~] 5.6 Error handling patterns in real codebases
- [~] 5.7 Project milestone: URL shortener core

## Part 2: Concurrency and the standard library

### Module 6: Concurrency fundamentals

- [~] 6.1 Goroutines and the cost model
- [~] 6.2 Channels: unbuffered, buffered, direction (ex)
- [~] 6.3 `select`, timeouts and `time.After` pitfalls
- [~] 6.4 `sync.WaitGroup`, `WaitGroup.Go`, `Mutex`, `RWMutex`, `Once` (ex)
- [~] 6.5 `sync/atomic` and atomic types
- [~] 6.6 The race detector
- [~] 6.7 The Go memory model, in plain words

### Module 7: Concurrency patterns and context

- [~] 7.1 `context.Context`: cancellation, deadlines, values (ex)
- [~] 7.2 Worker pools (ex)
- [~] 7.3 Pipelines and fan-out / fan-in
- [~] 7.4 `errgroup` and bounded concurrency
- [~] 7.5 Rate limiting and backpressure
- [~] 7.6 Goroutine leaks: finding and preventing them (`goroutineleak` profile)
- [~] 7.7 Graceful shutdown
- [~] 7.8 Projects: token bucket rate limiter and concurrent crawler

### Module 8: Standard library deep dives

- [~] 8.1 `io` and composition of readers and writers (ex)
- [~] 8.2 `encoding/json` and `encoding/json/v2`
- [~] 8.3 `time`: durations, monotonic clock, time zones
- [~] 8.4 `log/slog` structured logging
- [~] 8.5 `os`, `filepath`, `io/fs` and `embed`
- [~] 8.6 `strings`, `bytes`, `strconv`, `fmt` verbs
- [~] 8.7 `net/http` client done right (timeouts, transports)
- [~] 8.8 Project: mini Redis server

## Part 3: Quality, performance and internals

### Module 9: Testing in depth

- [~] 9.1 Table-driven tests and subtests (ex)
- [~] 9.2 Test helpers, `t.Cleanup`, `t.Context`, golden files
- [~] 9.3 Testable examples
- [~] 9.4 Mocks, fakes and interfaces for testing
- [~] 9.5 `httptest` (ex)
- [~] 9.6 Fuzzing (ex)
- [~] 9.7 Benchmarks with `b.Loop`
- [~] 9.8 Coverage and integration tests
- [~] 9.9 Testing concurrent code with `testing/synctest` (ex: mutation testing in 9.1)

### Module 10: Profiling and performance

- [~] 10.1 Measure first: benchmarks and `benchstat`
- [~] 10.2 CPU and memory profiling with `pprof`
- [~] 10.3 Execution tracer
- [~] 10.4 Allocation reduction and escape analysis
- [~] 10.5 Profile-guided optimization (PGO)
- [~] 10.6 `sync.Pool` and when it helps
- [~] 10.7 Exercise: optimize under an allocation budget (ex)

### Module 11: Runtime internals

- [~] 11.1 How the scheduler works (G, M, P)
- [~] 11.2 Stacks and goroutine growth
- [~] 11.3 The garbage collector, `GOGC` and `GOMEMLIMIT`
- [~] 11.4 Memory layout, alignment and `unsafe` (carefully)
- [~] 11.5 Interfaces and method dispatch under the hood
- [~] 11.6 Maps and slices under the hood
- [~] 11.7 Exercise: shrink a struct (ex)

## Part 4: Production Go

### Module 12: Building HTTP services

- [~] 12.1 `net/http` server and the enhanced `ServeMux` routing
- [~] 12.2 Middleware
- [~] 12.3 Request validation and JSON APIs
- [~] 12.4 Configuration and the twelve-factor app
- [~] 12.5 Observability: logs, metrics, traces (OpenTelemetry)
- [~] 12.6 Security basics: timeouts, limits, TLS, secrets
- [~] 12.7 Exercise: a notes JSON API (ex)
- [~] 12.8 Project milestone: shortener HTTP API

### Module 13: Working with data

- [~] 13.1 `database/sql` and connection pooling
- [~] 13.2 `pgx` and PostgreSQL
- [~] 13.3 Migrations
- [~] 13.4 Transactions and context
- [~] 13.5 Caching with Redis
- [~] 13.6 Exercise: a read-through cache (ex)
- [~] 13.7 Project milestone: shortener on PostgreSQL
- [~] 13.8 Project: a PostgreSQL job queue

### Module 14: CLIs and tooling

- [~] 14.1 `flag` vs Cobra
- [~] 14.2 Reading config, env and files
- [~] 14.3 Code generation with `go generate`
- [~] 14.4 Releasing binaries (GoReleaser, version stamping)
- [~] 14.5 Exercise: a to-do CLI (ex)
- [~] 14.6 Project: shortctl, a CLI client for the shortener

### Module 15: Go for cloud native

- [~] 15.1 Containerizing Go: multi-stage builds, distroless, static binaries
- [~] 15.2 Go in Kubernetes: probes, signals, graceful shutdown
- [~] 15.3 `client-go` basics
- [~] 15.4 Writing a Kubernetes controller with controller-runtime
- [~] 15.5 Operators in practice
- [~] 15.6 Exercise: drain gracefully (ex)
- [~] 15.7 Project milestone: the shortener on Kubernetes

## Part 5: Projects and case studies

See [case-studies.md](case-studies.md) for the detailed list and sourcing status.
