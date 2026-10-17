---
title: "3. Composite types"
weight: 3
description: "Arrays, slices, maps, structs, embedding and iterators: how Go groups data, and the memory behavior you need to understand to use them safely."
---

Slices, maps and structs are the data structures you will use in almost every Go program. They look familiar, but their **memory behavior** is different from lists, dictionaries and classes in other languages. This module explains what happens underneath, so the surprising cases stop being surprising.

**Time:** about 3 to 4 hours, plus exercises and the project.

{{< cards >}}
  {{< card link="01-arrays" title="3.1 Arrays" subtitle="Fixed size, value semantics, and why slices are used instead." >}}
  {{< card link="02-slices" title="3.2 Slices" subtitle="Length, capacity and the backing array." >}}
  {{< card link="03-slice-gotchas" title="3.3 Slice gotchas" subtitle="Aliasing, append in functions, memory retention." >}}
  {{< card link="04-slices-and-maps-packages" title="3.4 The slices and maps packages" subtitle="Sorting, searching, editing and cloning." >}}
  {{< card link="05-maps" title="3.5 Maps" subtitle="Lookups, sets, grouping and nil maps." >}}
  {{< card link="06-structs" title="3.6 Named types and structs" subtitle="Types, literals, tags and anonymous structs." >}}
  {{< card link="07-embedding" title="3.7 Embedding and composition" subtitle="Promotion, conflicts and why it is not inheritance." >}}
  {{< card link="08-iterators" title="3.8 Iterators" subtitle="range-over-func, iter.Seq and lazy pipelines." >}}
  {{< card link="09-checkpoint-project" title="3.9 Checkpoint: wordfreq" subtitle="Count words in files with maps, slices and sorting." >}}
{{< /cards >}}
