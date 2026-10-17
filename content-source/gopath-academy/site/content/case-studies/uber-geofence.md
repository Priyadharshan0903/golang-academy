---
title: "Uber: 170,000 queries per second with a geofence service"
linkTitle: "Uber: geofence service"
weight: 5
description: "Uber's geofence lookup service, written in Go in 2015, became its highest query-per-second service: an in-memory index behind a read-write lock, refreshed in the background, at under 5 ms p95 latency."
---

| | |
|---|---|
| **Company** | Uber |
| **System** | Geofence lookup: which city and zone a location is in |
| **Source** | Kai Wei, [How We Built Uber Engineering's Highest Query per Second Service Using Go](https://www.uber.com/blog/go-geofence-highest-query-per-second-service/), Uber Engineering blog, 24 February 2016 |
| **Connects to** | Modules 6, 7, 10, 12 |

## The story

Uber uses **geofences**, polygons on a map, for location-based configuration: where a product is available, airport rules, dynamic pricing areas. A lookup answers "which geofences contain this point?", and almost every request in the marketplace needs one.

The team's main language was Node.js, but in early 2015 they chose Go for this service, for three reasons given in the post:

- **Throughput and latency**: hundreds of thousands of queries per second, with a 99th percentile target under 100 ms.
- **CPU-heavy work**: point-in-polygon tests suit a compiled, statically typed language better than interpreted JavaScript.
- **Background refreshes**: in single-threaded Node.js, CPU-heavy refreshing of the data would stall queries; goroutines run refreshes on other cores while queries continue.

### The design

- **Every instance holds all the data in memory.** The service is stateless and not partitioned; a deterministic polling schedule keeps instances in sync, and data is also written to local disk for fast restarts.
- **A simple two-level index** instead of an R-tree or S2 cells: first find the city, then check the geofences in that city. It is still O(N), but N drops from tens of thousands to hundreds.
- **A read-write lock** protects the index. New data is built on the side and swapped in. The team first tried atomic pointer swaps (`StorePointer`/`LoadPointer`), found them brittle, and went back to the lock: slightly more latency, much simpler code.

### Results

On New Year's Eve 2015, the service peaked at **170,000 queries per second** in Uber's main non-China data center, on **40 machines at 35% CPU**, with latency **under 5 ms at the 95th percentile** and **under 50 ms at the 99th**. It had 99.99% uptime since launch; the post attributes the downtime to programming errors and a file descriptor leak in a third-party library, not to the Go runtime.

## What changed since

Atomic pointers are much less brittle today: `atomic.Pointer[T]` (Go 1.19) is type-safe, and swapping an immutable index through one is a common pattern (lesson 6.5). The read-write lock remains a perfectly good choice when, as here, measurements show it is fast enough.

## What you can apply

- **Simple beats clever when it is fast enough.** A two-level scan and a read-write lock served 170k QPS. Measure before reaching for complex indexes (Module 10).
- **Build new data on the side, then swap it in**, so readers never see a half-updated structure (lessons 6.4 and 6.5).
- **Keep CPU-heavy background work off the request path**: goroutines make that natural (Module 7).
- **Know your percentiles.** The targets and results are stated as p95 and p99, not averages (lesson 12.5).

## Sources

- Kai Wei: [How We Built Uber Engineering's Highest Query per Second Service Using Go](https://www.uber.com/blog/go-geofence-highest-query-per-second-service/), Uber Engineering blog, 24 February 2016
