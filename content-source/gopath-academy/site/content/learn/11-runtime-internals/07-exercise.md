---
title: "11.7 Exercise: shrink a struct"
linkTitle: "11.7 Exercise: shrink a struct"
weight: 7
description: "Apply what you learned about alignment and padding: reorder a struct's fields to save 25% of its memory, checked by unsafe.Sizeof."
goVersion: "1.27"
---

## The task

{{< exercise dir="m11/layout" >}}
`Event` is 64 bytes on a 64-bit platform. Reorder its fields so that `unsafe.Sizeof(Event{})` is 48 bytes. You may not add, remove, rename or change the type of any field; `TestFieldsUnchanged` checks that with reflection.
{{< /exercise >}}

```bash
cd exercises
go test ./m11/layout     # TestSize fails: 64 bytes
```

## Work it out on paper first

Before moving any field, write down each field's size and alignment (the table in lesson 11.4), then answer:

1. How many bytes of data does `Event` hold, ignoring padding?
2. What is the struct's alignment? (The largest alignment of any field.)
3. The size must be a multiple of the alignment. What is the smallest possible size?

<details>
<summary>Check your answers</summary>

1. 16 (string) + 8 + 8 + 4 + 2 + 1 + 1 + 1 = 41 bytes.
2. 8 bytes.
3. 41 rounded up to a multiple of 8 is **48**. So 48 is not just a target; it is the minimum.

</details>

<details>
<summary>Hint: where do the 64 bytes come from?</summary>

Use `unsafe.Offsetof` on each field of the starter, or reason it through: `Active` (1 byte) is followed by 7 bytes of padding so `ID` starts at offset 8; `Kind` by 7 more before `Score`; `Retries` by 6 before `Name`. That is 20 bytes of padding inside, plus 3 at the end.

</details>

The reference solution, with offsets in comments, is in `exercises/m11/layout/solution.go`.

## Going further

- Run the `fieldalignment` analyzer from lesson 11.4 on the starter and compare its suggestion with yours.
- Write a benchmark that allocates a `[]Event` of one million elements with each layout, and compare `B/op`.

## Part 3, second half: recap

You now know what happens underneath your code: how goroutines are scheduled onto threads and why `GOMAXPROCS` matters in containers, how stacks grow and move, how the garbage collector works and how `GOGC` and `GOMEMLIMIT` steer it, how values are laid out in memory, what interface and generic calls cost, and how slices and maps grow and hold on to memory.

Next is **Part 4: Production Go**, starting with **[Module 12: Building HTTP services](/learn/12-http-services/)**, where all of this comes together in real services.
