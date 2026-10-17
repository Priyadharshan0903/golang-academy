---
title: "2. Language basics"
weight: 2
description: "Variables, types, constants, strings, control flow, functions, defer, pointers and scope: the core of the language, including the gotchas."
---

This module covers the core of the language. You already know what a variable or a loop is, so the focus is on **how Go does it differently**, and on the gotchas that bite developers coming from other languages.

**Time:** about 3 hours, plus exercises.

{{< cards >}}
  {{< card link="01-variables" title="2.1 Variables and zero values" subtitle="var, :=, and why Go has no uninitialized variables." >}}
  {{< card link="02-basic-types" title="2.2 Basic types and conversions" subtitle="Integers, floats, explicit conversions and overflow." >}}
  {{< card link="03-constants-and-iota" title="2.3 Constants and iota" subtitle="Untyped constants, enums and bit flags." >}}
  {{< card link="04-strings-bytes-runes" title="2.4 Strings, bytes and runes" subtitle="UTF-8, len, range and building strings." >}}
  {{< card link="05-control-flow" title="2.5 Control flow" subtitle="if, for, switch and labels." >}}
  {{< card link="06-functions" title="2.6 Functions and closures" subtitle="Multiple returns, variadics, function values." >}}
  {{< card link="07-defer" title="2.7 defer" subtitle="Cleanup, evaluation order and traps." >}}
  {{< card link="08-pointers" title="2.8 Pointers without fear" subtitle="&, *, nil and when to use them." >}}
  {{< card link="09-scope-and-shadowing" title="2.9 Scope and shadowing" subtitle="The := bug every Go developer writes once." >}}
{{< /cards >}}
