---
title: "5. Errors"
weight: 5
description: "Errors as values, wrapping with context, sentinel errors and error types, joining multiple errors, panic and recover, and error handling patterns from real codebases."
---

Go has no exceptions. Functions return errors as ordinary values, and callers decide what to do with them. This module shows how to do that **well**: errors that carry useful context, that callers can inspect reliably, and that are handled in the right place.

**Time:** about 2 to 3 hours, plus exercises and the project milestone.

{{< cards >}}
  {{< card link="01-errors-are-values" title="5.1 Errors are values" subtitle="The error interface, creating and handling errors." >}}
  {{< card link="02-wrapping" title="5.2 Wrapping, errors.Is and errors.As" subtitle="Adding context while keeping errors inspectable." >}}
  {{< card link="03-sentinels-and-types" title="5.3 Sentinel errors vs error types" subtitle="Choosing what callers can depend on." >}}
  {{< card link="04-joining-errors" title="5.4 errors.Join and multi-errors" subtitle="Reporting several problems at once." >}}
  {{< card link="05-panic-and-recover" title="5.5 panic and recover" subtitle="When, and mostly when not." >}}
  {{< card link="06-patterns" title="5.6 Error handling patterns" subtitle="Where to handle, log, retry and translate errors." >}}
  {{< card link="07-checkpoint-project" title="5.7 Project: URL shortener core" subtitle="Milestone 1 of a project that grows through the course." >}}
{{< /cards >}}
