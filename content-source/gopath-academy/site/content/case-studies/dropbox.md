---
title: "Dropbox: moving performance-critical backends from Python to Go"
linkTitle: "Dropbox"
weight: 4
description: "Around 2013, Dropbox began moving its performance-critical backends from Python to Go for concurrency and speed, wrote about 200,000 lines of Go, and open sourced the libraries it had to build along the way."
---

| | |
|---|---|
| **Company** | Dropbox |
| **Change** | Performance-critical backends moved from Python to Go, starting around mid-2013 |
| **Source** | Patrick Lee, [Open Sourcing Our Go Libraries](https://dropbox.tech/infrastructure/open-sourcing-our-go-libraries), Dropbox Tech blog, 1 July 2014 |
| **Connects to** | Modules 1, 5, 7, 13 |

## The story

Python helped Dropbox grow and iterate quickly. As its user base grew, the company wanted to scale more efficiently, and about a year before this July 2014 post it decided to move its **performance-critical backends from Python to Go**, for better concurrency support and faster execution. A small team wrote about **200,000 lines of Go**, and by the time of the post major parts of the infrastructure had moved.

The obstacle they kept hitting was not the language but the ecosystem: Go was young, and robust libraries for building large systems were missing. So they wrote their own, and open sourced them as `godropbox`:

- **caching**: building blocks for caching layers;
- **errors**: errors that carry stack traces;
- **database/sqlbuilder**: generating SQL programmatically;
- **memcache**: a memcache client with connection pooling and sharding;
- **net2**: connection management;
- **hash2**: including a space-efficient consistent hashing algorithm.

The post does not name the specific systems that moved or give performance numbers, so neither does this case study.

## What changed since

The library gap Dropbox describes has largely closed. Several of the things it had to build are now in the standard library or are mature open source projects:

| Dropbox built (2014) | Today |
|---|---|
| errors with extra context | `%w` wrapping, `errors.Is`/`As`/`Join` (Module 5) |
| SQL helpers, connection pools | `database/sql` pooling, pgx, sqlc (Module 13) |
| connection management | `net/http` transport pooling, `context` deadlines (Modules 7 and 8) |
| caching layers | Redis clients, `singleflight` (lesson 13.5) |

## What you can apply

- **Move the hot paths, not everything.** Dropbox targeted performance-critical backends; most Python code was fine where it was. Measure first (Module 10).
- **Budget for libraries** when adopting a language in a big organization: the language is often the easy part.
- **Open sourcing internal libraries** let others improve them, and Dropbox committed to using the public repository internally so fixes landed there first.

## Sources

- Patrick Lee: [Open Sourcing Our Go Libraries](https://dropbox.tech/infrastructure/open-sourcing-our-go-libraries), Dropbox Tech blog, 1 July 2014
