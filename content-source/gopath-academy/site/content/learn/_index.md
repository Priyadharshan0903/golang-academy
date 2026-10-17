---
title: Learn Go
linkTitle: Learn
weight: 1
cascade:
  type: docs
---

This course takes you from your first Go program to production services and runtime internals. It is written for developers who already know at least one programming language.

## How the course works

1. **Read the lesson.** Each one explains a concept, shows why it matters, and walks through tested examples.
2. **Run the examples.** Clone the repository and run any example yourself. Every example on the site comes from a real file in the [`examples/`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/examples) folder.
3. **Do the exercise.** Most lessons end with an exercise in [`exercises/`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/exercises). Run `go test` and make the tests pass.
4. **Build the checkpoint project** at the end of each module.

```bash
git clone https://github.com/Priyadharshan0903/gopath-academy.git
cd gopath-academy
go test ./examples/...
```

## Modules

{{< cards >}}
  {{< card link="01-setup-and-tooling" title="1. Setup and tooling" subtitle="Install Go, write your first program, modules, packages, the go command, formatting and linting." icon="terminal" >}}
  {{< card link="02-language-basics" title="2. Language basics" subtitle="Variables, types, constants, strings, control flow, functions, defer, pointers and scope." icon="code" >}}
  {{< card link="03-composite-types" title="3. Composite types" subtitle="Arrays, slices, maps, structs, embedding and iterators, with a word counting project." icon="collection" >}}
  {{< card link="04-methods-interfaces-generics" title="4. Methods, interfaces and generics" subtitle="Methods, implicit interfaces, generics and Go 1.27 generic methods, plus API design." icon="puzzle" >}}
  {{< card link="05-errors" title="5. Errors" subtitle="Errors as values, wrapping, sentinels and types, errors.Join, panic and recover, plus the URL shortener core." icon="exclamation-circle" >}}
  {{< card link="06-concurrency-fundamentals" title="6. Concurrency fundamentals" subtitle="Goroutines, channels, select, sync, atomics, the race detector and the memory model." icon="lightning-bolt" >}}
  {{< card link="07-concurrency-patterns" title="7. Concurrency patterns and context" subtitle="context, worker pools, pipelines, errgroup, rate limiting, leaks, shutdown, plus two projects." icon="switch-horizontal" >}}
  {{< card link="08-standard-library" title="8. Standard library deep dives" subtitle="io, JSON and json/v2, time, slog, files and embed, formatting, HTTP clients, plus a mini Redis." icon="library" >}}
  {{< card link="09-testing" title="9. Testing in depth" subtitle="Table tests, helpers, examples, fakes, httptest, fuzzing, b.Loop benchmarks, coverage, synctest." icon="beaker" >}}
  {{< card link="10-performance" title="10. Profiling and performance" subtitle="benchstat, pprof, the tracer, escape analysis, PGO and sync.Pool, all measured." icon="chart-bar" >}}
  {{< card link="11-runtime-internals" title="11. Runtime internals" subtitle="Scheduler, stacks, GC tuning, memory layout, interfaces, slices and maps, with experiments." icon="chip" >}}
  {{< card link="12-http-services" title="12. Building HTTP services" subtitle="ServeMux routing, middleware, JSON APIs, config, observability and security." icon="server" >}}
  {{< card link="13-working-with-data" title="13. Working with data" subtitle="PostgreSQL with database/sql and pgx, migrations, transactions, Redis caching." icon="database" >}}
  {{< card link="14-clis-and-tooling" title="14. CLIs and tooling" subtitle="flag and Cobra, layered config, go generate, reproducible releases." icon="terminal" >}}
  {{< card link="15-cloud-native" title="15. Go for cloud native" subtitle="Containers, Kubernetes rollouts, client-go and a real operator." icon="cloud" >}}
{{< /cards >}}

More modules are on the way. The full outline is in the project's [curriculum plan](https://github.com/Priyadharshan0903/gopath-academy/blob/main/plan/curriculum.md).

## Which Go version?

The course is written for **Go 1.27**, the current stable release. Example code builds on **Go 1.24 or newer**. When a lesson relies on a newer feature, it says which version introduced it.
