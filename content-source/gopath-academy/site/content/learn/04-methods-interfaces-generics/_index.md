---
title: "4. Methods, interfaces and generics"
weight: 4
description: "Attach behavior to types with methods, write flexible code with implicit interfaces, and remove duplication with generics, including Go 1.27 generic methods."
---

This module is where Go's design really shows. There are no classes and no inheritance; instead you have **methods** on any named type, **interfaces** that are satisfied implicitly, and **generics** for code that works across types. Together they replace most of what class hierarchies do in other languages, with less coupling.

**Time:** about 4 hours, plus exercises.

{{< cards >}}
  {{< card link="01-methods" title="4.1 Methods" subtitle="Value vs pointer receivers, nil receivers, method values." >}}
  {{< card link="02-interfaces" title="4.2 Interfaces" subtitle="Implicit satisfaction, small interfaces, method sets." >}}
  {{< card link="03-any-and-type-switches" title="4.3 any, assertions and type switches" subtitle="Inspecting dynamic types safely." >}}
  {{< card link="04-nil-interfaces" title="4.4 Nil interfaces" subtitle="The interface that holds a nil pointer." >}}
  {{< card link="05-standard-interfaces" title="4.5 Standard interfaces" subtitle="Stringer, io.Reader, io.Writer, sort.Interface." >}}
  {{< card link="06-generics" title="4.6 Generics" subtitle="Type parameters, constraints and generic types." >}}
  {{< card link="07-generic-methods" title="4.7 Generic methods and limits" subtitle="Go 1.27 generic methods, and when not to use generics." >}}
  {{< card link="08-api-design" title="4.8 Designing APIs" subtitle="Accept interfaces, return structs, options." >}}
{{< /cards >}}
