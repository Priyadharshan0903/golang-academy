---
title: "8.6 strings, bytes, strconv and fmt verbs"
linkTitle: "8.6 strings, strconv, fmt"
weight: 6
description: "The fmt verbs you need, fast number and string conversion with strconv, the strings helpers that save you loops, aligned table output with text/tabwriter, and allocation-free formatting."
goVersion: "1.27"
---

## What you will learn

- The `fmt` verbs, widths and flags you will use most
- `strconv` for conversion and quoting, including allocation-free `Append` functions
- `strings` helpers: `Cut`, `CutLast` (Go 1.27), `NewReplacer`, `EqualFold`, `Map`
- Aligned, column-based output with `text/tabwriter`
- `fmt.Appendf` for building output without extra allocations

## `fmt` verbs

{{< example file="examples/m08/formatting/formatting_test.go" region="verbs" run="go test -run Example_verbs -v ./examples/m08/formatting" >}}

| Verb | Meaning |
|---|---|
| `%v`, `%+v`, `%#v` | default format; with field names; as Go syntax |
| `%T` | the type |
| `%d`, `%x`, `%o`, `%b` | integer in decimal, hex, octal, binary |
| `%f`, `%.2f`, `%e`, `%g` | float: decimal, 2 decimals, scientific, compact |
| `%s`, `%q` | string; quoted and escaped |
| `%c`, `%U` | rune as character; as `U+XXXX` |
| `%t` | boolean |
| `%w` | error wrapping, only in `fmt.Errorf` (lesson 5.2) |
| `%%` | a literal percent sign |

Flags and widths go between `%` and the verb: `%5d` pads to 5 columns, `%-5d` left-aligns, `%05d` pads with zeros, `%8.3f` is width 8 with 3 decimals, and `%*d` takes the width from an argument.

`go vet` checks every `Printf`-style call: wrong verb for the type, missing or extra arguments. Keep it in CI.

## `strconv`

{{< example file="examples/m08/formatting/formatting_test.go" region="strconv" run="go test -run Example_strconv -v ./examples/m08/formatting" >}}

- `Atoi`, `ParseInt`, `ParseFloat`, `ParseBool` parse text and return errors you must check (lesson 2.2).
- `Itoa`, `FormatInt(n, base)`, `FormatFloat` convert to text, faster than `fmt.Sprint`.
- `Quote` and `Unquote` handle Go-style quoted strings.
- `AppendInt`, `AppendFloat`, `AppendQuote` write into an existing `[]byte`. In hot paths (encoders, log formatters) they avoid allocating a string for every number.

## `strings` helpers

{{< example file="examples/m08/formatting/formatting_test.go" region="strings" run="go test -run Example_stringsHelpers -v ./examples/m08/formatting" >}}

- `strings.Cut(s, sep)` splits around the **first** separator and reports whether it was found. It replaces most `Index` plus slicing code and many `SplitN(s, sep, 2)` calls.
- `strings.NewReplacer` performs many replacements in one pass, and is safe for concurrent use, so build it once.
- `strings.EqualFold` compares case-insensitively without allocating. It uses simple Unicode case folding, so `"Straße"` and `"STRASSE"` are not equal; full language-aware comparison needs `golang.org/x/text`.
- `strings.Map` transforms or drops each rune.

`strings.Title` is deprecated because it handles many languages incorrectly; use `golang.org/x/text/cases` instead.

### `strings.CutLast` (Go 1.27)

{{< example file="examples/m08/formatting/cutlast_test.go" region="cutlast" run="go test -run Example_cutLast -v ./examples/m08/formatting" >}}

`CutLast` splits around the **last** separator: file extensions, the port in `host:port` (including IPv6 hosts with colons), the last path element. `bytes.CutLast` does the same for byte slices.

## Aligned output with `tabwriter`

{{< example file="examples/m08/formatting/formatting_test.go" region="tabwriter" run="go test -run Example_tabwriter -v ./examples/m08/formatting" >}}

`text/tabwriter` turns tab-separated cells into aligned columns, the way `kubectl get pods` prints its tables. Write rows with `\t` between cells, then call `Flush`.

## Formatting without extra allocations

{{< example file="examples/m08/formatting/formatting_test.go" region="append" run="go test -run Example_appendf -v ./examples/m08/formatting" >}}

`fmt.Appendf`, `fmt.Append` and `fmt.Appendln` (Go 1.19) format into a byte slice you provide, instead of returning a new string. Reusing one buffer across many lines saves allocations. Like the `strconv.Append` functions, this matters only in hot paths; elsewhere, `Sprintf` is perfectly fine.

## Common mistakes

### `fmt.Sprintf("%d", n)` in hot loops

`strconv.Itoa(n)` is several times faster and allocates less. Measure before optimizing, but know the option exists.

### Building SQL or shell commands with `Sprintf`

String formatting is not escaping. Use query parameters (lesson 13.1) and `exec.Command` with separate arguments.

### Ignoring `go vet` format warnings

A mismatched verb prints garbage like `%!d(string=three)` in production logs.

## Key takeaways

- Learn `%v`, `%+v`, `%T`, `%q`, widths and precision; let `go vet` check format strings.
- Use `strconv` for fast conversions, and its `Append` functions in hot paths.
- `strings.Cut` and `CutLast` (Go 1.27) replace manual index arithmetic.
- `text/tabwriter` aligns columns; `fmt.Appendf` formats into reusable buffers.

## Further reading

- [Package fmt](https://pkg.go.dev/fmt), especially the printing verbs section
- [Package strconv](https://pkg.go.dev/strconv) and [strings](https://pkg.go.dev/strings)
- [Package text/tabwriter](https://pkg.go.dev/text/tabwriter)
