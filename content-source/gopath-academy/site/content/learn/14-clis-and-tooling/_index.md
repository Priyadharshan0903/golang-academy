---
title: "14. CLIs and tooling"
weight: 14
description: "Command-line tools in Go: subcommands with flag and Cobra, layered configuration, code generation with go generate and go tool, and reproducible, versioned releases with GoReleaser."
---

Go is the language of command-line tools: `kubectl`, `docker`, `terraform`, `gh` and `hugo` are all Go programs. A single static binary that starts instantly, cross-compiles to every platform with one environment variable, and needs no runtime installed: that is hard to beat for tools.

This module covers what makes a CLI pleasant to use and easy to ship: a clear command structure, sensible configuration, generated code where it saves effort, and releases people can trust. It ends with **shortctl**, a client for the shortener API.

**Time:** about 3 to 4 hours, plus the exercise and project.

{{< cards >}}
  {{< card link="01-flag-and-cobra" title="14.1 flag vs Cobra" subtitle="Subcommands, I/O, exit codes, testability." >}}
  {{< card link="02-configuration" title="14.2 Configuration for CLIs" subtitle="Defaults, files, environment, flags." >}}
  {{< card link="03-go-generate" title="14.3 Code generation" subtitle="go generate, go tool, and your own generator." >}}
  {{< card link="04-releasing" title="14.4 Releasing binaries" subtitle="Versions, reproducible builds, GoReleaser." >}}
  {{< card link="05-exercise" title="14.5 Exercise: a to-do CLI" subtitle="Subcommands and exit codes from a spec." >}}
  {{< card link="06-project-shortctl" title="14.6 Project: shortctl" subtitle="A CLI client for the shortener API." >}}
{{< /cards >}}
