---
title: "dl.google.com: replacing a C++ download server with Go"
linkTitle: "dl.google.com"
weight: 2
description: "Google's download server for Chrome, Android SDK and Earth was rewritten from C++ to Go in 2012: incrementally, verified against live traffic, with the standard library doing most of the work and less than half the code."
---

| | |
|---|---|
| **Company** | Google |
| **System** | dl.google.com, the HTTP download server for Chrome, the Android SDK, Google Earth and more |
| **Source** | Brad Fitzpatrick, [dl.google.com: Powered by Go](https://go.dev/talks/2013/oscon-dl.slide), OSCON, 26 July 2013 |
| **Connects to** | Modules 4, 8, 9, 12 |

## The story

By 2012, dl.google.com was an old C++ server built around a single-threaded event loop. At startup it copied every payload to local disk, which by then took 12 to 24 hours. The code had duplicated copy-and-paste paths, incomplete documentation and tests, and threads added ad hoc. Brad Fitzpatrick's talk attributes its random crashes and stalls to that complexity.

The Go rewrite was done **incrementally**. First only the payload server was ported. Both versions ran in an **audit mode** where live traffic was mirrored to the old and the new server and their responses compared, so the new one was proven on real requests before it took over.

Much of the work disappeared into the standard library. `net/http` and `http.ServeContent` provided Range requests, ETags and `Last-Modified` handling, features the C++ server had implemented, incorrectly, in several places. The payload logic was built from small interfaces composed together: a type combining a size with `io.ReaderAt`, a multi-reader that concatenates parts, and a reader that issues 2 MB aligned reads so they match the cache's chunks. Caching used **groupcache**, Fitzpatrick's open-source library, which coordinates a cache across many servers and makes sure a miss is fetched only once instead of by every server at the same moment.

## Results

According to the talk, the rewrite removed 39 files, 14,032 lines of C++ and Python, leaving a single binary, and the overall result was "less than half the code". The new server used the same CPU for the same bandwidth but could serve much more bandwidth on multiple cores, used less memory and no local disk, started instantly instead of after hours, handled download spikes without crashing, and came with tests.

## What you can apply

- **Lean on the standard library.** Serving files correctly (ranges, conditional requests, content types) is subtle; `http.ServeContent` already gets it right (Module 12).
- **Small interfaces compose.** `io.ReaderAt` plus a few wrappers replaced a lot of special cases (lessons 4.x and 8.1).
- **Migrate incrementally, verify with real traffic.** Mirroring requests to old and new implementations and comparing responses is a powerful way to rewrite a running system.
- **Prevent duplicate work on cache misses.** groupcache's idea is the same as `singleflight` and request coalescing (lesson 13.5).

## Sources

- Brad Fitzpatrick: [dl.google.com: Powered by Go](https://go.dev/talks/2013/oscon-dl.slide), OSCON, 26 July 2013
