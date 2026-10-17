---
title: "3.9 Checkpoint project: wordfreq"
linkTitle: "3.9 Checkpoint: wordfreq"
weight: 9
description: "Build a command line tool that counts word frequencies across files or standard input, using maps, slices, sorting and bufio."
goVersion: "1.27"
---

## What you will build

`wordfreq` reads text from files or standard input and prints the most frequent words:

```bash
go run ./projects/wordfreq/cmd/wordfreq -n 4 projects/wordfreq/internal/count/testdata/sample.txt
```

```text
      4  go
      3  is
      1  an
      1  and
```

It uses everything from this module: a map to count, slices and `slices.SortStableFunc` to rank, a struct for results, plus `bufio.Scanner` to read line by line (a preview of Module 8).

The code is in [`projects/wordfreq`](https://github.com/Priyadharshan0903/gopath-academy/tree/main/projects/wordfreq).

## The design

{{% steps %}}

### Split the problem

Two pure functions do the real work, so they are easy to test:

- `count.Words(r io.Reader) (map[string]int, error)` turns text into counts.
- `count.Top(counts, n) []Entry` turns counts into a ranked list.

`main` only handles flags, files and printing.

### Count words

{{< example file="projects/wordfreq/internal/count/count.go" >}}

Things to notice:

- `bufio.Scanner` reads one line at a time, so even huge files use little memory. Its default limit is 64 KiB per line; `sc.Buffer` raises it.
- `strings.FieldsFunc` splits on every rune that is not part of a word.
- `words := fields[:0]` filters **in place**, the pattern from lesson 3.3.
- `Top` first sorts keys alphabetically, then does a **stable** sort by count. Stability keeps the alphabetical order among equal counts, which makes the output deterministic despite random map order.

### A bug the tests found

The first version only treated letters and digits as part of a word. A test with Tamil text failed: `வணக்கம்` was counted as two words.

In scripts such as Tamil, Hindi or Bengali, vowel signs and the virama are **combining marks** (`unicode.IsMark`), not letters. Treating them as separators breaks words apart. This is the grapheme lesson from 2.4 showing up in real code, and it is why the test file includes non-English input.

### Test it

{{< example file="projects/wordfreq/internal/count/count_test.go" >}}

### Wire up the command

{{< example file="projects/wordfreq/cmd/wordfreq/main.go" >}}

`addFile` is a separate function so that each file's `defer f.Close()` runs right after that file is processed, avoiding the `defer` in a loop trap from lesson 2.7.

{{% /steps %}}

## Try it

```bash
cd projects/wordfreq
go test ./...
cat ../../site/content/learn/*/*.md | go run ./cmd/wordfreq -n 15
```

## Stretch goals

1. Add a `-min` flag to ignore words shorter than N runes (use `utf8.RuneCountInString`, not `len`).
2. Add a `-stop` flag with a file of stop words to ignore (a set from lesson 3.5).
3. Return an iterator from a new `count.WordsSeq(r io.Reader) iter.Seq[string]` and rebuild `Words` on top of it.
4. Print the output as JSON with a `-json` flag (a preview of lesson 8.2).

## Module 3 recap

You can now:

- Explain the difference between arrays and slices, and predict when slices share memory
- Avoid the common `append` and memory retention traps
- Use maps idiomatically for lookups, sets, counters and grouping
- Model data with named types, structs and tags
- Compose types with embedding, without mistaking it for inheritance
- Write and consume iterators

Next is **[Module 4: Methods, interfaces and generics](/learn/04-methods-interfaces-generics/)**, where these types get behavior.
