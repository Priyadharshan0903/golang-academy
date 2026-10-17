---
title: "10.7 Exercise: optimize under an allocation budget"
linkTitle: "10.7 Exercise: optimize"
weight: 7
description: "Take correct but wasteful code, profile it, and make it pass an allocation budget test without breaking correctness."
goVersion: "1.27"
---

## The task

This exercise is the module in miniature. The starter code is **correct**, and its correctness tests already pass. But it allocates wildly, and a test enforces an allocation budget it misses by a factor of about 50.

{{< exercise dir="m10/hashtags" >}}
`TopTags` finds the most common hashtags in a list of posts. Make `TestAllocationBudget` pass (at most 400 allocations for 1000 posts) while keeping `TestTopTags` and `TestTiesAreAlphabetical` green. Use the workflow from this module: run the benchmark, read a memory profile, fix the biggest source, measure again.
{{< /exercise >}}

## Suggested workflow

```bash
cd exercises

# 1. Where do we stand?
go test ./m10/hashtags                                  # TestAllocationBudget fails
go test -run xxx -bench . -benchmem ./m10/hashtags      # note ns/op, B/op, allocs/op

# 2. Where do the allocations come from?
go test -run xxx -bench . -memprofile mem.out ./m10/hashtags
go tool pprof -top -sample_index=alloc_objects mem.out
go tool pprof -list TopTags -sample_index=alloc_objects mem.out

# 3. Fix the biggest source, then repeat steps 1 and 2.
```

On the test VM, the starter measured about **19,840 allocations** and 760 µs per call; the reference solution measured **255 allocations** and about 95 µs.

## Hints

Try to find these yourself in the profile before reading them.

<details>
<summary>Hint 1: where the starter allocates</summary>

Four places: `strings.Fields` allocates a slice for every post; `tag += string(r)` allocates a new string for every character; `strings.ToLower` allocates even when the tag is already lowercase; and the sort comparator calls `fmt.Sprintf` twice on every comparison.

</details>

<details>
<summary>Hint 2: substrings are free</summary>

A substring like `post[i+1:j]` shares memory with `post` and does not allocate. Using it as a map key does not allocate either. Scan the bytes of each post directly instead of splitting it into words.

</details>

<details>
<summary>Hint 3: sorting</summary>

`slices.SortFunc` with `cmp.Or(cmp.Compare(...), strings.Compare(...))` (lesson 3.4) compares without allocating.

</details>

The reference solution is in `exercises/m10/hashtags/solution.go`.

## Part 3, first half: recap

You can now measure performance honestly with benchmarks and benchstat, find hot spots with CPU and memory profiles, investigate latency with the tracer and flight recorder, reason about allocations with escape analysis, apply PGO, and reuse objects with `sync.Pool`, always checking the result with data.

Next is **[Module 11: Runtime internals](/learn/11-runtime-internals/)**, which explains the machinery behind what you have been measuring.
