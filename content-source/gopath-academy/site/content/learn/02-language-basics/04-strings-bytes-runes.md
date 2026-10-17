---
title: "2.4 Strings, bytes, runes and UTF-8"
linkTitle: "2.4 Strings and runes"
weight: 4
description: "How Go strings really work: immutable byte sequences holding UTF-8, the difference between bytes and runes, and how to build and process text efficiently."
goVersion: "1.27"
---

## What you will learn

- What a Go string is under the hood
- Why `len("héllo")` is not 5
- The difference between a `byte` and a `rune`, and when each one matters
- How `range` decodes UTF-8 for you
- How to build strings efficiently and the most useful `strings` functions

## The idea

A Go string is an **immutable sequence of bytes**. By convention (and in all Go source code) those bytes are **UTF-8** encoded text, but the string type itself does not enforce that.

UTF-8 uses 1 byte for ASCII characters and 2 to 4 bytes for everything else. So there are two different ways to count a string:

- **Bytes**: what `len(s)` returns, and what `s[i]` indexes.
- **Runes**: Unicode code points, which is closer to what people mean by "characters". A `rune` is an `int32` holding one code point.

For plain English text the two counts are the same, which is why bugs here often go unnoticed until someone types `é`, `ü`, `த` or an emoji.

## Bytes vs runes

{{< example file="examples/m02/text/text_test.go" region="bytes" run="go test -run Example_bytesVsRunes -v ./examples/m02/text" >}}

`"héllo, 世界"` has 9 characters but takes 14 bytes: `é` is 2 bytes and each Chinese character is 3. `s[1]` returns the **first byte** of `é` (195), not the character.

Slicing (`s[i:j]`) also works on byte offsets. Cutting in the middle of a multi-byte character gives you invalid UTF-8.

## `range` decodes UTF-8

A `for ... range` loop over a string is special: it walks the string **one rune at a time**, giving you the byte offset and the decoded rune:

{{< example file="examples/m02/text/text_test.go" region="range" run="go test -run Example_rangeOverString -v ./examples/m02/text" >}}

Notice the offsets jump from 1 to 3, because `é` occupied bytes 1 and 2. If the string contains invalid UTF-8, `range` yields `U+FFFD` (the replacement character `�`) for each bad byte instead of failing.

Use `range` whenever you process text character by character. Use byte indexing only when you know the data is ASCII or you really want bytes (parsing a protocol, for example).

## Converting between strings, bytes and runes

{{< example file="examples/m02/text/text_test.go" region="convert" run="go test -run Example_conversions -v ./examples/m02/text" >}}

- `[]byte(s)` gives a **mutable copy** of the bytes. Use it for I/O and binary data.
- `[]rune(s)` decodes into one element per code point. Use it when you need random access by character, such as reversing text.
- `string(b)` and `string(r)` convert back, again copying.

Strings are immutable: you cannot do `s[0] = 'N'`. You build a new string instead.

## Building strings

`s += "x"` in a loop creates a new string on every iteration. For more than a handful of pieces, use `strings.Builder`, which grows one buffer:

{{< example file="examples/m02/text/text_test.go" region="builder" run="go test -run Example_builder -v ./examples/m02/text" >}}

The zero value of `strings.Builder` is ready to use (the zero value rule from lesson 2.1 again), and `fmt.Fprintf` can write straight into it.

The functions you will reach for most often in the `strings` package:

| Need | Function |
|---|---|
| Contains, prefix, suffix | `strings.Contains`, `HasPrefix`, `HasSuffix` |
| Find position | `strings.Index`, `LastIndex` |
| Split and join | `strings.Split`, `Fields`, `Join` |
| Split once around a separator | `strings.Cut` |
| Trim | `strings.TrimSpace`, `Trim`, `TrimPrefix`, `TrimSuffix` |
| Case | `strings.ToLower`, `ToUpper`, `EqualFold` (case-insensitive compare) |
| Replace | `strings.ReplaceAll`, `strings.NewReplacer` |
| Repeat | `strings.Repeat` |

The `bytes` package has the same functions for `[]byte`, and `unicode` has rune classification such as `unicode.IsLetter` and `unicode.IsSpace`.

## Raw string literals

Backquotes create a **raw string**: no escape sequences, and it can span lines. Ideal for regular expressions, file paths on Windows, SQL and embedded templates.

{{< example file="examples/m02/text/text_test.go" region="raw" run="go test -run Example_rawStrings -v ./examples/m02/text" >}}

## Going deeper: runes are not always characters

A rune is a code point, and some things a person sees as one character are several code points. `é` can be written as one code point (U+00E9) or as `e` followed by a combining accent (U+0301). Many emoji, flags and family emoji are sequences of several code points. These user-perceived characters are called **grapheme clusters**.

For most programs, runes are the right level. If you are building something like a text editor or truncating user names for display, use a package that handles grapheme clusters (for example `golang.org/x/text` for normalization, or a third party segmentation library).

## Common mistakes

### Using `len` as a character count

`len` counts bytes. For characters, use `utf8.RuneCountInString(s)`. For a display width limit, think about grapheme clusters too.

### Indexing to get "the nth character"

`s[n]` is the nth **byte**. Convert to `[]rune` first, or loop with `range` and count.

### Concatenating in a loop

`result += piece` inside a loop is quadratic for large inputs. Use `strings.Builder` or `strings.Join`.

### Comparing case-insensitively with `ToLower`

`strings.ToLower(a) == strings.ToLower(b)` allocates two new strings and is not correct for every language. `strings.EqualFold(a, b)` is both faster and more correct.

## Exercise

{{< exercise dir="m02/palindrome" >}}
Implement `Reverse` so it works for any UTF-8 text, and `IsPalindrome`, which ignores case and anything that is not a letter or digit. The tests include accented letters and Chinese characters, so byte-based solutions will fail.
{{< /exercise >}}

## Key takeaways

- A string is an immutable sequence of bytes, normally UTF-8.
- `len` and indexing work in bytes; `range` and `[]rune` work in runes.
- `[]byte(s)` and `[]rune(s)` copy. Strings themselves can never be modified.
- Build strings with `strings.Builder` or `strings.Join`, not `+=` in a loop.
- Runes are code points, not always what a user sees as one character.

## Further reading

- [Strings, bytes, runes and characters in Go](https://go.dev/blog/strings) on the Go blog
- [Package strings](https://pkg.go.dev/strings), [unicode/utf8](https://pkg.go.dev/unicode/utf8) and [unicode](https://pkg.go.dev/unicode)
- [For statements with range clause](https://go.dev/ref/spec#For_range) in the spec
