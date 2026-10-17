---
title: "11. Runtime internals"
weight: 11
description: "How the Go runtime actually works: the scheduler, goroutine stacks, the garbage collector and its knobs, memory layout, interfaces, slices and maps, each shown with experiments you can run."
---

Module 10 taught you to measure. This module explains **why** the measurements come out the way they do. You do not need any of this to write correct Go, but it is what lets you predict performance, read profiles with understanding, and configure services well, especially in containers.

Every claim in this module is backed by a small experiment in `examples/m11`. Run them: several give different answers on different Go versions, and those differences are part of the lesson.

**Time:** about 3 to 4 hours, plus the exercise.

{{< cards >}}
  {{< card link="01-scheduler" title="11.1 The scheduler: G, M and P" subtitle="How goroutines are multiplexed onto threads." >}}
  {{< card link="02-stacks" title="11.2 Goroutine stacks" subtitle="Small, growable, and they move." >}}
  {{< card link="03-gc" title="11.3 The garbage collector" subtitle="GOGC, GOMEMLIMIT, Green Tea, weak pointers." >}}
  {{< card link="04-memory-layout" title="11.4 Memory layout and unsafe" subtitle="Sizes, alignment, padding and false sharing." >}}
  {{< card link="05-interfaces" title="11.5 Interfaces under the hood" subtitle="Two words, itabs, and the cost of dispatch." >}}
  {{< card link="06-slices-and-maps" title="11.6 Slices and maps under the hood" subtitle="Growth, Swiss tables, and memory that stays." >}}
  {{< card link="07-exercise" title="11.7 Exercise: shrink a struct" subtitle="Save 25% of memory by reordering fields." >}}
{{< /cards >}}
