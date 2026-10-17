---
title: "1.1 Why Go, and when not to use it"
linkTitle: "1.1 Why Go"
weight: 1
description: "What Go was designed for, where it shines, and the cases where another language is the better choice."
goVersion: "1.27"
---

## What you will learn

- The problems Go was designed to solve
- The handful of design choices that make Go feel different from Java, Python or JavaScript
- Where Go is a strong choice, and where it honestly is not
- How Go releases and its compatibility promise work

## The idea

Go is a statically typed, compiled, garbage collected language with built-in concurrency. It was designed at Google, starting in 2007, by Robert Griesemer, Rob Pike and Ken Thompson. It was open sourced in 2009 and reached version 1.0 in 2012.

The designers were frustrated by a specific set of problems in large codebases: very slow builds, complicated dependency management, languages that had grown too many features, and the difficulty of writing correct concurrent servers. Go is their answer to those problems, which is why it often feels like it is missing things on purpose.

A useful way to think about Go: **it optimizes for the team and the codebase over years, not for the individual writing clever code today.**

## What makes Go feel different

| Design choice | What it means in practice |
|---|---|
| Small language | The spec is short enough to read in an afternoon. There is usually one obvious way to write something. |
| Fast compilation | Large projects build in seconds, so the edit, build, test loop stays tight. |
| Single static binary | `go build` produces one executable with no runtime to install. Deploying is copying a file. |
| Cross-compiling built in | Build a Linux binary from a Mac with two environment variables. No extra toolchain. |
| Goroutines and channels | Concurrency is part of the language. Starting thousands of goroutines is normal. |
| Garbage collector tuned for low latency | You do not manage memory by hand, and GC pauses are typically very short. |
| Composition over inheritance | No classes, no inheritance. Structs, methods and small interfaces that are satisfied implicitly. |
| Errors are values | No exceptions. Functions return an `error`, and you handle it where it happens. |
| One formatting style | `gofmt` formats all Go code the same way. Nobody argues about braces. |
| Strong standard library | HTTP servers, JSON, crypto, testing and profiling are all in the box. |

If you come from **Java**, expect fewer abstractions and no framework that dominates everything. If you come from **Python or JavaScript**, expect a compiler that catches far more mistakes and code that runs much faster, at the cost of a bit more typing. If you come from **Rust or C++**, expect a much simpler language that gives up some control (a garbage collector, no manual memory layout tricks by default) in exchange for speed of development.

## Where Go is a strong choice

- **Cloud infrastructure and platform tooling.** Docker, Kubernetes, Terraform, Prometheus and etcd are all written in Go. If you work on infrastructure, you will read Go code whether you choose to or not.
- **Network services and APIs.** The standard library HTTP server is production ready, and goroutines make handling many concurrent connections straightforward.
- **Command line tools.** Single binaries that start instantly and cross-compile easily are ideal for CLIs.
- **Concurrent data processing.** Pipelines, workers and fan-out patterns are natural to express.
- **Large teams.** Simple, uniform code is easier to review and to hand over.

## Where another language may fit better

Being honest about this is part of learning a language well.

- **Hard real-time or ultra-low-latency systems.** Go's GC pauses are short, but they exist. If a few hundred microseconds of jitter is unacceptable, consider Rust, C or C++.
- **Data science and machine learning.** The Python ecosystem (NumPy, pandas, PyTorch) has no real equivalent in Go.
- **Desktop GUI applications.** Possible, but the ecosystem is thin compared to other platforms.
- **Very small embedded devices.** Standard Go binaries are several megabytes. TinyGo exists for microcontrollers, but it is a separate compiler with limitations.
- **Problems that want a rich type system.** Go has generics, but it deliberately lacks features like sum types or pattern matching. If you want to encode lots of rules in types, Rust, Scala or Haskell will feel more natural.

We will look at real examples of both sides, including a company that moved a service away from Go, in the [case studies](/case-studies/) (see [Discord](/case-studies/discord/)).

## Releases and the compatibility promise

Two facts shape how you work with Go over time:

1. **A new major release ships every six months**, in February and August. Each release is supported until two newer major releases exist, so you get about a year of security fixes per version.
2. **The Go 1 compatibility promise.** Code written for Go 1.0 should still compile and run with the latest Go 1.x release. New features are added carefully, and old code keeps working. This is why ten year old Go tutorials still mostly work, and why upgrading Go is usually uneventful.

{{< callout type="info" >}}
This course is written for **Go 1.27**. Thanks to the compatibility promise, almost everything here also works on older 1.2x versions. When a feature is newer, the lesson says which version introduced it.
{{< /callout >}}

## Common mistakes

### Writing Go like your previous language

The most common problem for experienced developers is not syntax. It is importing patterns that fight the language: deep inheritance hierarchies rebuilt with embedding, getters and setters for every field, an interface for every struct "just in case", or exceptions simulated with `panic`. Each lesson in this course calls out these habits where they appear.

### Expecting a framework to decide everything

Many languages have one dominant framework. In Go, the standard library is the framework for a surprising amount of work. Reach for a third party library when it clearly earns its place, not by default.

## Key takeaways

- Go was designed for large codebases and teams: fast builds, simple code, easy deployment and built-in concurrency.
- It is a natural fit for infrastructure, network services and CLIs.
- It is a weaker fit for hard real-time systems, data science, GUIs and tiny embedded devices.
- New releases come every six months, and the compatibility promise means upgrades rarely break code.

## Further reading

- [Go at Google: Language Design in the Service of Software Engineering](https://go.dev/talks/2012/splash.article), Rob Pike's explanation of why Go looks the way it does
- [Frequently Asked Questions (FAQ)](https://go.dev/doc/faq) on go.dev
- [Go 1 and the Future of Go Programs](https://go.dev/doc/go1compat), the compatibility promise
- [Go release history](https://go.dev/doc/devel/release)
