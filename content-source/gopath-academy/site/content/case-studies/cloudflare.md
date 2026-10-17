---
title: "Cloudflare: early production Go at the edge"
linkTitle: "Cloudflare"
weight: 3
description: "By 2013 Cloudflare ran its DNS proxy, its Railgun compression system, a two-person-rule encryption server and certificate tooling in Go, relying on a goroutine per request and recovered panics for isolation."
---

| | |
|---|---|
| **Company** | Cloudflare |
| **Systems** | RRDNS (DNS proxy), Railgun, Red October, SSL Bundler |
| **Source** | John Graham-Cumming, [What we've been doing with Go](https://blog.cloudflare.com/what-weve-been-doing-with-go/), Cloudflare blog, 11 November 2013 |
| **Connects to** | Modules 4, 5, 6, 12 |

## The story

Cloudflare was one of the earliest companies to run Go in production. In July 2012 its only Go project was **Railgun**, a system that compresses traffic between Cloudflare's data centers and customers' origin servers, written entirely in Go. Sixteen months later, John Graham-Cumming listed several more:

- **RRDNS**, a DNS proxy built to scale Cloudflare's DNS infrastructure;
- **Red October**, an encryption server that enforces a **two-person rule**: decrypting sensitive material such as private keys requires at least two authorized people to be logged in;
- **SSL Bundler**, which computes the fastest, shortest intermediate certificate chain for a customer's certificate;
- internal load-testing tools, and experiments with Docker and CoreOS's Raft library.

The post explains which Go features mattered in RRDNS:

- **A goroutine per request**, which isolates requests from each other.
- **`defer` and `recover`**, so a panic while handling one request is recovered and logged instead of crashing the whole DNS server.
- **Garbage collection**, which avoided complex reference counting.
- **Goroutines or read-write locks** for safe access to shared data.
- **Interfaces** that keep modules self-contained while allowing very different implementations behind them, some using cgo, some using background workers.

In the [Go blog's fourth birthday post](https://go.dev/blog/4years) the same year, Graham-Cumming summarized: Go was "the perfect match for our needs."

## What changed since

The panic-recovery pattern RRDNS describes is now built into `net/http`, which recovers handler panics itself; lesson 12.2 adds a middleware that turns them into logged 500 responses with request IDs. Cloudflare has since written many more Go services, and its engineering blog became a well-known source of Go production advice, such as its guide to `net/http` timeouts referenced in lesson 12.6.

## What you can apply

- **Isolate requests.** One goroutine per request, with panics recovered at the boundary, keeps one bad input from taking down a server (lessons 5.5 and 12.2).
- **Prefer simple concurrency tools for shared state.** A read-write lock is often all you need (lesson 6.4).
- **Hide implementation details behind small interfaces** so modules can evolve independently (Module 4).

## Sources

- John Graham-Cumming: [What we've been doing with Go](https://blog.cloudflare.com/what-weve-been-doing-with-go/), Cloudflare blog, 11 November 2013
- Andrew Gerrand: [Four years of Go](https://go.dev/blog/4years), The Go Blog, 10 November 2013
