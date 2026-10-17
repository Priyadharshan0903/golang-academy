---
title: "Kubernetes and Docker: why cloud native was written in Go"
linkTitle: "Kubernetes and Docker"
weight: 1
description: "Docker (2013) and Kubernetes (2014) were both written in Go. Their authors explained why: static binaries, concurrency, tooling, and a language neutral enough to attract contributors, plus the pain points they hit early."
---

| | |
|---|---|
| **Projects** | Docker (first released 2013), Kubernetes (open sourced 2014) |
| **Sources** | Docker team slides, November 2013; Joe Beda, Gopher Academy blog, November 2014; Go blog, November 2013 |
| **Connects to** | Modules 1, 6, 14, 15 |

## The story

The two projects that defined cloud native infrastructure were both written in Go, at a time when Go was only a few years old.

**Docker.** In a November 2013 talk, Docker engineers Michael Crosby, Jérôme Petazzoni and Victor Vieux listed their reasons ([slides](https://www.slideshare.net/slideshow/docker-and-go-why-did-we-decide-to-write-docker-in-go/28015076)):

- **Static compilation**: `go build` produces a binary that embeds almost everything it needs, which matters for a tool that must run on any Linux host.
- **Neutrality**: Go was not C++, Python, Ruby or Java, so no existing community "owned" the project.
- **The right features**: asynchronous primitives, low-level interfaces to processes and system calls, a large standard library.
- **Tooling**: documentation, formatting, testing and dependency fetching built into the `go` command.
- **Multi-architecture builds** with file name suffixes such as `_linux.go` instead of a preprocessor.

The same slides were candid about the pain: maps are not safe for concurrent use, `go get` could not pin versions (Docker had to vendor its dependencies), tests had no cleanup hooks, the `flag` package lacked short and long options, and error handling was verbose. In the [Go blog's fourth birthday post](https://go.dev/blog/4years), Docker's creator Solomon Hykes put it more bluntly: had Docker not been written in Go, "it would not have been as successful."

**Kubernetes.** Joe Beda, one of Kubernetes' creators, wrote that Go was "the only logical choice" for a project started from scratch in 2014 ([Kubernetes + Go = Crazy Delicious](https://blog.gopheracademy.com/birthday-bash-2014/kubernetes-go-crazy-delicious/)). His reasons: a language "neither too high level nor too low level", good standard and third-party libraries, fast builds and tests, a culture of simple code, tools such as `gofmt`, `go vet`, coverage and the race detector, easy fan-out of network calls, and memory safety when parsing untrusted input. He also explained the alternatives the team rejected: C and C++ as daunting for casual contributors, Java because of its heavy runtime install, and Python because dynamic typing is risky for systems software.

## What changed since

Almost every pain point in the Docker slides has since been addressed by Go itself:

| 2013 pain point | Today |
|---|---|
| `go get` cannot pin versions; vendor everything | Go modules with `go.mod` and `go.sum` (lesson 1.4) |
| No test cleanup | `t.Cleanup` (Go 1.14), `TestMain`, `t.Context` (Go 1.24) (Module 9) |
| `flag` lacks short and long options | still true; Cobra and pflag, used by both projects (lesson 14.1) |
| No IDE | gopls, the language server behind every major editor (lesson 1.2) |
| Maps not safe for concurrent use | still by design; `sync.Mutex`, `sync.Map`, and the race detector (Module 6) |

The pull of these projects also worked in the other direction: because Kubernetes, Docker, containerd, etcd and Prometheus are written in Go, the libraries for extending them (client-go, controller-runtime) are Go libraries, which is a large part of why platform teams write Go today.

## What you can apply

- **Static binaries make distribution trivial.** That is still the main reason infrastructure tools choose Go; lessons 14.4 and 15.1 build on it.
- **Choose boring, neutral tools for projects that need contributors.** Simplicity of the code base was a feature for both projects.
- **Fan-out with goroutines** is the everyday pattern of distributed systems code (Module 7).
- **Write Go that drives Kubernetes**: client-go and controller-runtime exist because Kubernetes is Go (lessons 15.3 and 15.4).

## Sources

- Michael Crosby, Jérôme Petazzoni, Victor Vieux: [Docker and Go: why did we decide to write Docker in Go?](https://www.slideshare.net/slideshow/docker-and-go-why-did-we-decide-to-write-docker-in-go/28015076), November 2013
- Andrew Gerrand: [Four years of Go](https://go.dev/blog/4years), The Go Blog, 10 November 2013
- Joe Beda: [Kubernetes + Go = Crazy Delicious](https://blog.gopheracademy.com/birthday-bash-2014/kubernetes-go-crazy-delicious/), Gopher Academy Blog, 11 November 2014
