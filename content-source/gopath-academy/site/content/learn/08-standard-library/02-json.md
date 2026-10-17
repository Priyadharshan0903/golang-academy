---
title: "8.2 encoding/json and encoding/json/v2"
linkTitle: "8.2 JSON"
weight: 2
description: "Encode and decode JSON with struct tags, stream with Encoder and Decoder, customize types, handle unknown and polymorphic data, and use the encoding/json/v2 API that became generally available in Go 1.27."
goVersion: "1.27"
---

## What you will learn

- How struct fields and tags map to JSON
- Decoding into structs, and what happens with missing, extra or mistyped fields
- Strict decoding and streaming with `Decoder` and `Encoder`
- Custom encoding with `MarshalText` and `MarshalJSON`
- Delaying decoding with `json.RawMessage`, and the trouble with numbers in `any`
- `omitzero` (Go 1.24) versus `omitempty`
- The `encoding/json/v2` and `jsontext` packages (Go 1.27)

## The idea

`encoding/json` uses **reflection** to map Go values to JSON and back. Exported struct fields become JSON object members; **tags** control the names and options. There is no code generation and no schema: the struct is the schema.

{{< callout type="info" >}}
**What changed in Go 1.27.** The new packages `encoding/json/v2` and `encoding/json/jsontext` are now generally available, and the classic `encoding/json` package is now implemented **on top of v2**. Its behavior is preserved and you do not need to migrate, but **error message text may differ** from older releases. Never match on JSON error strings; check error types instead, as the examples do.
{{< /callout >}}

## Encoding

{{< example file="examples/m08/jsonenc/jsonenc_test.go" region="marshal" run="go test -run Example_marshal -v ./examples/m08/jsonenc" >}}

| Tag | Effect |
|---|---|
| `json:"name"` | Use `name` as the member name |
| `json:"name,omitempty"` | Leave out if the value is empty: `0`, `""`, `false`, `nil`, empty slice or map |
| `json:"name,omitzero"` | (Go 1.24+) Leave out if the value is the zero value, or its `IsZero()` method returns true |
| `json:"-"` | Never encode or decode this field |
| `json:"name,string"` | Encode a number or bool as a JSON string |

Unexported fields (`internal`) are invisible to the encoder. Pointer fields (`Address *Address`) encode as `null` or are omitted with `omitempty` when nil, which is the usual way to make a nested object optional. `time.Time` encodes as RFC 3339.

## Decoding

{{< example file="examples/m08/jsonenc/jsonenc_test.go" region="unmarshal" run="go test -run Example_unmarshal -v ./examples/m08/jsonenc" >}}

- Pass a **pointer**; `Unmarshal` fills the value it points to.
- Member names match field names **case-insensitively** in `encoding/json`.
- Members with no matching field (`extra`) are **ignored**.
- Fields missing from the JSON keep whatever value they had (usually the zero value). If you need to know whether a field was present, use a pointer field.
- A type mismatch returns a `*json.UnmarshalTypeError` with the field name. Inspect it with `errors.As`.

## Strict decoding

{{< example file="examples/m08/jsonenc/jsonenc_test.go" region="strict" run="go test -run Example_strictDecoding -v ./examples/m08/jsonenc" >}}

For APIs and config files, silently ignoring unknown fields hides typos (`nmae`). `Decoder.DisallowUnknownFields()` turns them into errors.

## Streaming

{{< example file="examples/m08/jsonenc/jsonenc_test.go" region="stream" run="go test -run Example_streaming -v ./examples/m08/jsonenc" >}}

`json.NewDecoder(r)` reads values from any `io.Reader` one at a time, so it handles large inputs and streams of concatenated values (newline-delimited JSON). `json.NewEncoder(w)` writes values to any `io.Writer`. Use them directly on HTTP request and response bodies instead of reading the body into a byte slice first.

## Custom encoding

{{< example file="examples/m08/jsonenc/jsonenc_test.go" region="custom" run="go test -run Example_customEncoding -v ./examples/m08/jsonenc" >}}

Implement `encoding.TextMarshaler` and `TextUnmarshaler` when a type has a natural text form (enums, IDs, IP addresses). JSON then encodes it as a string, it works as a **map key**, and other encoders (YAML, XML, flags) understand it too. Implement `json.Marshaler` / `json.Unmarshaler` (`MarshalJSON`, `UnmarshalJSON`) only when you need full control over the JSON structure.

Note the receivers: `MarshalText` on the value, `UnmarshalText` on the pointer, because unmarshaling must modify the value.

## Polymorphic data with `RawMessage`

{{< example file="examples/m08/jsonenc/jsonenc_test.go" region="raw" run="go test -run Example_rawMessage -v ./examples/m08/jsonenc" >}}

`json.RawMessage` keeps a member's raw bytes undecoded. Decode the envelope first, look at the discriminator (`type`), then decode the payload into the right struct. This is the standard way to handle event streams and webhooks with several payload shapes.

## Numbers in `any`

{{< example file="examples/m08/jsonenc/jsonenc_test.go" region="numbers" run="go test -run Example_numbersInAny -v ./examples/m08/jsonenc" >}}

Decoding into `any` turns **every** number into `float64`, which silently loses precision for integers above 2<sup>53</sup> (IDs from other systems are a common victim). Decode into typed struct fields, or call `Decoder.UseNumber()` to get `json.Number` values that keep the original text.

## `omitzero` vs `omitempty`

{{< example file="examples/m08/jsonenc/jsonenc_test.go" region="omitzero" run="go test -run Example_omitZero -v ./examples/m08/jsonenc" >}}

`omitempty` never omits a struct value, so an unset `time.Time` appears as `"0001-01-01T00:00:00Z"`. `omitzero` (Go 1.24+) omits zero values, and respects an `IsZero()` method, which is what you usually want for times and other structs.

## `encoding/json/v2` (Go 1.27)

{{< callout type="info" >}}
The examples below need **Go 1.27 or newer**; their file is gated with `//go:build go1.27`.
{{< /callout >}}

The v2 API takes **options** as extra arguments instead of requiring a `Decoder` for every setting, and has stricter, more predictable defaults:

{{< example file="examples/m08/jsonv2/jsonv2_test.go" region="basics" run="go test -run Example_v2Defaults -v ./examples/m08/jsonv2" >}}

{{< example file="examples/m08/jsonv2/jsonv2_test.go" region="strict" run="go test -run Example_v2Strictness -v ./examples/m08/jsonv2" >}}

Differences from v1 worth knowing:

| | v1 (`encoding/json`) | v2 (`encoding/json/v2`) |
|---|---|---|
| nil slice or map | `null` | `[]` or `{}` |
| Member name matching | case-insensitive | case-sensitive (option `MatchCaseInsensitiveNames`) |
| Duplicate member names | last one wins | error |
| Invalid UTF-8 in strings | replaced silently | error |
| Unknown members | ignored (`DisallowUnknownFields` on a `Decoder`) | ignored (option `RejectUnknownMembers`) |
| Streaming | `NewEncoder`, `NewDecoder` | `MarshalWrite`, `UnmarshalRead`, or `jsontext` |

The low-level `jsontext` package reads and writes JSON tokens, always producing valid JSON:

{{< example file="examples/m08/jsonv2/jsonv2_test.go" region="jsontext" run="go test -run Example_jsontext -v ./examples/m08/jsonv2" >}}

**Which should you use?** For existing code, keep `encoding/json`: it is supported, and on Go 1.27 it already runs on the faster v2 engine (unmarshaling in particular is significantly faster). For new code on Go 1.27+, v2's stricter defaults and option-based API are a good choice. Check the duplicate-name and case-sensitivity differences before switching an existing API, since clients may depend on the old behavior.

## Common mistakes

### Unexported fields

`name string` is never encoded or decoded. Capitalize it and add a tag.

### Unmarshaling into a non-pointer

`json.Unmarshal(data, u)` instead of `&u` returns an `InvalidUnmarshalError` and changes nothing.

### Relying on case-insensitive matching

It works in v1 and fails in v2. Use exact names in tags and in JSON.

### Matching error strings

The text of JSON errors changed in Go 1.27. Use `errors.As` with `*json.UnmarshalTypeError`, `*json.SyntaxError` and friends.

### Large integers through `any`

Precision loss above 2<sup>53</sup>. Use typed fields or `UseNumber`.

## Key takeaways

- Exported fields plus tags define the JSON shape; `-`, `omitempty`, `omitzero` and `string` are the main options.
- Decoding ignores unknown members unless you ask for strictness; type mismatches return typed errors.
- Use `Decoder` and `Encoder` to stream from and to readers and writers.
- Implement `MarshalText`/`UnmarshalText` for types with a text form; use `RawMessage` for polymorphic payloads.
- Go 1.27: `encoding/json/v2` and `jsontext` are available, and `encoding/json` runs on the v2 engine with the same behavior.

## Further reading

- [Package encoding/json](https://pkg.go.dev/encoding/json)
- [Package encoding/json/v2](https://pkg.go.dev/encoding/json/v2) and [encoding/json/jsontext](https://pkg.go.dev/encoding/json/jsontext)
- [Go 1.27 release notes](https://go.dev/doc/go1.27)
- [JSON and Go](https://go.dev/blog/json) on the Go blog
