---
title: "6. Concurrency fundamentals"
weight: 6
description: "Goroutines, channels, select, the sync and atomic packages, the race detector and the Go memory model."
---

Concurrency is where Go stands out, and also where the hardest bugs live. This module builds the foundations carefully: what goroutines cost, how channels synchronize, when a mutex is the better tool, how to find data races, and what the memory model actually guarantees.

**Time:** about 4 hours, plus exercises. Run every example and exercise in this module with `-race`.

{{< cards >}}
  {{< card link="01-goroutines" title="6.1 Goroutines" subtitle="Starting them, waiting for them, and what they cost." >}}
  {{< card link="02-channels" title="6.2 Channels" subtitle="Unbuffered, buffered, closing and direction." >}}
  {{< card link="03-select-and-timeouts" title="6.3 select and timeouts" subtitle="Waiting on several things, timers and tickers." >}}
  {{< card link="04-sync-package" title="6.4 The sync package" subtitle="WaitGroup, Mutex, RWMutex, Once." >}}
  {{< card link="05-atomics" title="6.5 sync/atomic" subtitle="Lock-free counters, flags and snapshots." >}}
  {{< card link="06-race-detector" title="6.6 The race detector" subtitle="Finding data races before production does." >}}
  {{< card link="07-memory-model" title="6.7 The memory model" subtitle="What happens-before means, in plain words." >}}
{{< /cards >}}
