---
title: "8. Standard library deep dives"
weight: 8
description: "io and bufio, encoding/json and json/v2, time, log/slog, files and embed, formatting, and a correct net/http client, then a mini Redis server built only on the standard library."
---

Go's standard library is large, consistent and production quality. Many Go services use few third party packages because the standard library already covers so much. This module goes deep on the packages you will use every week.

**Time:** about 4 to 5 hours, plus exercises and the project.

{{< cards >}}
  {{< card link="01-io" title="8.1 io and bufio" subtitle="Composing readers and writers, scanning, buffering." >}}
  {{< card link="02-json" title="8.2 encoding/json and json/v2" subtitle="Tags, streaming, custom encoding, the v2 API." >}}
  {{< card link="03-time" title="8.3 time" subtitle="Durations, layouts, time zones, monotonic clocks." >}}
  {{< card link="04-slog" title="8.4 log/slog" subtitle="Structured logging, handlers, levels, redaction." >}}
  {{< card link="05-files" title="8.5 Files, io/fs and embed" subtitle="Reading, atomic writes, walking, os.Root." >}}
  {{< card link="06-formatting" title="8.6 strings, strconv and fmt" subtitle="Verbs, conversions, aligned output." >}}
  {{< card link="07-http-client" title="8.7 net/http client done right" subtitle="Timeouts, bodies, status codes, transports." >}}
  {{< card link="08-project-miniredis" title="8.8 Project: mini Redis" subtitle="A TCP server that redis-cli can talk to." >}}
{{< /cards >}}
