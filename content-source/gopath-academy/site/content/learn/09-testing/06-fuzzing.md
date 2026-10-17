---
title: "9.6 Fuzzing"
linkTitle: "9.6 Fuzzing"
weight: 6
description: "Use Go's built-in fuzzer to find inputs that break your code: writing fuzz targets, choosing properties to check, seed corpora, saved failures and running fuzzing in CI."
goVersion: "1.27"
---

## What you will learn

- What fuzzing is and what kinds of bugs it finds
- How to write a fuzz target with `testing.F`
- **Properties** worth checking: round trips, no panics, agreement with a reference
- The seed corpus and the `testdata/fuzz` directory
- How to run the fuzzer and use what it finds

## The idea

Unit tests check the cases **you** thought of. A fuzzer generates **millions** of inputs automatically, guided by code coverage: when an input reaches new code, the fuzzer mutates it further. It is very good at finding the cases nobody thought of: empty input, huge input, invalid UTF-8, a dangling backslash, a length prefix that lies.

Since Go 1.18, fuzzing is built into `go test`.

## A fuzz target

Code to fuzz:

{{< example file="examples/m09/fuzzing/escape.go" region="code" >}}

Fuzz targets:

{{< example file="examples/m09/fuzzing/escape_test.go" region="fuzz" run="go test -run xxx -fuzz=FuzzRoundTrip -fuzztime=10s ./examples/m09/fuzzing" >}}

- A fuzz target is `func FuzzXxx(f *testing.F)` in a test file.
- `f.Add(...)` adds **seed** inputs. They should be representative valid inputs.
- `f.Fuzz(func(t *testing.T, args...))` is called with each generated input. Argument types can be `[]byte`, `string`, integers, floats, `bool` and `rune`.

## What to check

The hard part of fuzzing is not generating inputs but deciding what is **wrong** for an arbitrary input. Useful properties:

| Property | Example |
|---|---|
| **No panics, no hangs** on any input | parsers, decoders, protocol readers |
| **Round trip** | `Unquote(Quote(s)) == s`, `Decode(Encode(x)) == x` |
| **Agreement with a reference** | a fast implementation matches a slow, obviously correct one |
| **Invariants** | output is always valid UTF-8, sorted, within bounds |
| **Idempotence** | `Normalize(Normalize(s)) == Normalize(s)` |

The mini Redis RESP parser (lesson 8.8) has a "no panics" fuzz target, and the exercise below uses round trips.

## Running the fuzzer

```bash
go test ./examples/m09/fuzzing                              # seeds only: fast, runs in CI
go test -run xxx -fuzz=FuzzRoundTrip ./examples/m09/fuzzing  # fuzz until a failure or Ctrl+C
go test -run xxx -fuzz=FuzzRoundTrip -fuzztime=30s ./...     # fuzz for a fixed time
```

- Without `-fuzz`, fuzz targets run like normal tests over the seed corpus, so they cost almost nothing in CI.
- With `-fuzz`, one target runs for as long as you allow, on all CPU cores. `-run xxx` skips the regular tests.
- Inputs that found new code paths are cached in your Go build cache, so later fuzzing sessions start where you left off.

## When the fuzzer finds a bug

The failing input is written to `testdata/fuzz/FuzzXxx/<hash>` in the package directory, and from then on **plain `go test` runs it** as a regression test. The workflow:

1. Fuzz, get a failure.
2. Commit the new `testdata/fuzz` file.
3. Fix the code; `go test` now passes and keeps passing.

## Fuzzing in CI

Run seed corpora in every CI build (they are just tests). Run actual fuzzing on a schedule, for example nightly for a few minutes per target, or continuously with a service like OSS-Fuzz, which supports native Go fuzzing.

## Common mistakes

### Fuzzing without a property

A target that only calls the function catches panics, which is useful, but most bugs produce wrong output rather than panics. Add a property.

### Non-deterministic targets

A target that depends on time, randomness or global state produces failures the fuzzer cannot reproduce.

### Slow targets

The fuzzer's power comes from speed. Avoid network calls and big allocations inside the target.

## Exercise

{{< exercise dir="m09/rle" >}}
Implement a run-length `Encode` and `Decode`. The tests include two fuzz targets: a round-trip property and a "never panic, and anything accepted re-encodes consistently" property. After the unit tests pass, run `go test -fuzz=FuzzRoundTrip -fuzztime=20s ./m09/rle` and see whether the fuzzer can break your code (runs longer than 255 are a common trap).
{{< /exercise >}}

## Key takeaways

- Fuzzing generates inputs automatically, guided by coverage, and finds the cases you did not think of.
- Write `FuzzXxx(f *testing.F)` with seeds from `f.Add` and a property in `f.Fuzz`.
- Good properties: no panics, round trips, agreement with a reference, invariants.
- Plain `go test` runs seeds and saved failures; `go test -fuzz` explores.
- Commit failing inputs from `testdata/fuzz` as regression tests.

## Further reading

- [Go Fuzzing](https://go.dev/doc/security/fuzz/) on go.dev
- [Tutorial: Getting started with fuzzing](https://go.dev/doc/tutorial/fuzz)
