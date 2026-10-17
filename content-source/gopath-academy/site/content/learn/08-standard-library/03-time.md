---
title: "8.3 time: durations, monotonic clock, time zones"
linkTitle: "8.3 time"
weight: 3
description: "Work with time.Duration, format and parse with Go's reference layout, handle time zones correctly, measure elapsed time with the monotonic clock, and avoid the classic date arithmetic surprises."
goVersion: "1.27"
---

## What you will learn

- How `time.Duration` works and how to convert to and from it
- Go's unusual formatting layouts, based on a reference time
- Time zones: `In`, `FixedZone`, `LoadLocation`, `ParseInLocation`
- Why you compare instants with `Equal`, not `==`
- The monotonic clock and measuring elapsed time
- Date arithmetic surprises with `AddDate`

## Durations

{{< example file="examples/m08/timepkg/time_test.go" region="duration" run="go test -run Example_durations -v ./examples/m08/timepkg" >}}

A `time.Duration` is an `int64` count of **nanoseconds** with methods. Constants like `time.Second` make durations readable: `90 * time.Minute`. Two rules:

- Multiplying a constant by a duration works (`3 * time.Second`), but an `int` **variable** needs a conversion: `time.Duration(n) * time.Second`.
- Never pass plain numbers as durations (`time.Sleep(5)` sleeps 5 **nanoseconds**). Always multiply by a unit.

Durations print in a readable form (`1h30m0s`) and parse with `time.ParseDuration`, which is what the `flag` package uses for duration flags.

## Formatting and parsing layouts

Go does not use `YYYY-MM-DD` style patterns. A layout is the **reference time** written the way you want your output to look:

```text
Mon Jan 2 15:04:05 MST 2006
```

It reads as 1 2 3 4 5 6 7: month 1, day 2, hour 3 (15 in 24-hour), minute 4, second 5, year 2006, time zone -07.

{{< example file="examples/m08/timepkg/time_test.go" region="layout" run="go test -run Example_formatAndParse -v ./examples/m08/timepkg" >}}

| Want | Layout |
|---|---|
| Year | `2006` (or `06`) |
| Month | `01`, `1`, `Jan`, `January` |
| Day | `02`, `2`, `_2` |
| Hour | `15` (24h), `03` or `3` (12h) with `PM` |
| Minute, second | `04`, `05` |
| Fractional seconds | `.000`, `.000000`, `.999` (trailing zeros trimmed) |
| Time zone | `MST`, `-0700`, `-07:00`, `Z07:00` |

Use the predefined constants (`time.RFC3339`, `time.RFC3339Nano`, `time.DateOnly`, `time.DateTime`, `time.Kitchen`) whenever they fit. For APIs and logs, **RFC 3339 in UTC** is the safe default.

## Time zones

A `time.Time` is an **instant** plus a **location** used for display. Changing the location with `In` does not change the instant:

{{< example file="examples/m08/timepkg/time_test.go" region="zones" run="go test -run Example_timeZones -v ./examples/m08/timepkg" >}}

- `t.Equal(u)` compares instants. `==` compares the whole struct, including the location and the monotonic reading, so two equal instants in different zones are `!=`. Use `Equal` (and `Before`, `After`, `Compare`).
- `time.Parse` without an offset in the input assumes **UTC**. Use `time.ParseInLocation` to interpret a local wall-clock time in a specific zone.
- `time.FixedZone` creates a zone with a constant offset. For real zones with daylight saving rules, use `time.LoadLocation("Asia/Kolkata")` or `"America/New_York"`. It reads the system's time zone database; in minimal containers that database may be missing, so add `import _ "time/tzdata"` to embed it in your binary (about 450 KB).

A good rule for services: **store and compute in UTC, convert to a user's zone only for display.**

## Measuring elapsed time

{{< example file="examples/m08/timepkg/time_test.go" region="monotonic" run="go test -run Example_measuring -v ./examples/m08/timepkg" >}}

The wall clock can jump: NTP corrections, daylight saving changes, an admin setting the clock. `time.Now()` also records a **monotonic** clock reading that only moves forward, and `time.Since`, `time.Until` and `t.Sub` use it automatically when both times have one. So measuring durations with `time.Now()` and `time.Since()` is safe. The monotonic reading is stripped by `t.Round(0)` and by serialization, so a time read back from JSON or a database has none.

## Date arithmetic

{{< example file="examples/m08/timepkg/time_test.go" region="arithmetic" run="go test -run Example_arithmetic -v ./examples/m08/timepkg" >}}

- `Add` adds an exact duration.
- `AddDate(years, months, days)` works on the calendar, and **normalizes overflow**: January 31 plus one month is "February 31", which becomes March 3. If you need "the last day of next month", compute it explicitly.
- Around daylight saving transitions, a calendar day is not always 24 hours. `AddDate(0, 0, 1)` keeps the wall-clock time; `Add(24 * time.Hour)` does not.
- `t.Unix()` and `time.Unix(sec, nsec)` convert to and from Unix timestamps.

## Testing code that uses time

Code that calls `time.Now()` directly is hard to test. Two approaches from earlier lessons:

- **Inject a clock**: a `now func() time.Time` field, replaced in tests (the URL shortener and rate limiter projects do this).
- **`testing/synctest`** (Go 1.25+): run the test in a bubble with a fake clock that advances only when all goroutines are blocked (lesson 6.3). Go 1.27 adds `synctest.Sleep`, which sleeps and then waits for the bubble to settle.

## Common mistakes

### `time.Sleep(5)`

Five nanoseconds. Write `5 * time.Second`.

### Comparing times with `==`

Fails for equal instants in different zones, and when one time has a monotonic reading and the other does not. Use `Equal`.

### Parsing local times as UTC

`time.Parse("2006-01-02 15:04", "...")` assumes UTC. Use `ParseInLocation` when the input is local time.

### Assuming the time zone database exists

`LoadLocation` fails in `scratch` or distroless images without `tzdata`. Embed it with `import _ "time/tzdata"`.

### Using wall-clock differences for timeouts

Computing timeouts from two wall-clock readings breaks when the clock jumps. Use `time.Since`, timers, or context deadlines.

## Key takeaways

- `Duration` is nanoseconds; always multiply by a unit, and convert int variables explicitly.
- Layouts are the reference time `Mon Jan 2 15:04:05 MST 2006`; prefer the predefined constants, especially RFC 3339.
- A `Time` is an instant plus a display location; compare with `Equal`, store in UTC.
- `time.Now` carries a monotonic reading, so `time.Since` is safe for measuring.
- `AddDate` normalizes overflowing dates; check edge cases.

## Further reading

- [Package time](https://pkg.go.dev/time)
- [Package time/tzdata](https://pkg.go.dev/time/tzdata)
- [Monotonic clocks](https://pkg.go.dev/time#hdr-Monotonic_Clocks) in the time package documentation
