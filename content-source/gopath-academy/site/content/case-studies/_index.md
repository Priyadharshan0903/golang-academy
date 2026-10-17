---
title: Case studies
weight: 3
cascade:
  type: docs
---

How real companies and projects use Go in production: why they chose it, what went well, what did not, and which lessons in this course the story connects to. Every claim links to a primary source: an engineering blog post, a conference talk, a research paper, or the official Go blog. Where a source gives no number, neither do we.

{{< cards >}}
  {{< card link="kubernetes-and-docker" title="Kubernetes and Docker" subtitle="Why cloud native infrastructure was written in Go." >}}
  {{< card link="dl-google-com" title="dl.google.com" subtitle="Replacing a C++ download server with less than half the code." >}}
  {{< card link="cloudflare" title="Cloudflare" subtitle="Goroutine per request, recovered panics, early production Go." >}}
  {{< card link="dropbox" title="Dropbox" subtitle="Moving performance-critical backends from Python to Go." >}}
  {{< card link="uber-geofence" title="Uber: geofence service" subtitle="170k queries per second with a read-write lock." >}}
  {{< card link="uber-data-races" title="Uber: 2,000 data races" subtitle="What running the race detector on 46 million lines found." >}}
  {{< card link="twitch" title="Twitch: GC tuning" subtitle="A 10 GiB ballast, and why GOMEMLIMIT replaced it." >}}
  {{< card link="discord" title="Discord: Go to Rust" subtitle="When the garbage collector was the wrong fit." >}}
  {{< card link="monzo" title="Monzo" subtitle="A bank built as Go microservices on Kubernetes." >}}
{{< /cards >}}

## How to read them

Most of these stories are years old, and Go has changed a lot since: the garbage collector, the scheduler, modules, generics. Each case study notes what has changed since it was written, so you can tell which lessons still apply as stated and which were solved by later releases.

## More stories

The Go team collects case studies at [go.dev/solutions](https://go.dev/solutions/), including American Express, PayPal, Microsoft, Riot Games and many more.
