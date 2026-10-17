package timepkg

import (
	"fmt"
	"time"
)

// [start duration]

func Example_durations() {
	d := 90 * time.Minute
	fmt.Println(d, d.Hours())

	timeout := 2*time.Second + 500*time.Millisecond
	fmt.Println(timeout, timeout.Milliseconds())

	parsed, _ := time.ParseDuration("1h15m30s")
	fmt.Println(parsed.Round(time.Hour), parsed.Truncate(time.Hour))

	n := 3
	fmt.Println(time.Duration(n) * time.Second) // convert ints explicitly

	// Output:
	// 1h30m0s 1.5
	// 2.5s 2500
	// 1h0m0s 1h0m0s
	// 3s
}

// [end duration]

// [start layout]

func Example_formatAndParse() {
	t := time.Date(2026, time.October, 16, 14, 5, 9, 0, time.UTC)

	// Layouts are written using the reference time Mon Jan 2 15:04:05 MST 2006.
	fmt.Println(t.Format("2006-01-02"))
	fmt.Println(t.Format("02 Jan 2006, 3:04 PM"))
	fmt.Println(t.Format(time.RFC3339))

	p, err := time.Parse("2006-01-02 15:04", "2026-12-31 23:59")
	fmt.Println(p, err)

	// Output:
	// 2026-10-16
	// 16 Oct 2026, 2:05 PM
	// 2026-10-16T14:05:09Z
	// 2026-12-31 23:59:00 +0000 UTC <nil>
}

// [end layout]

// [start zones]

func Example_timeZones() {
	ist := time.FixedZone("IST", 5*3600+30*60) // UTC+05:30
	utc := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)

	local := utc.In(ist) // same instant, different presentation
	fmt.Println(local)
	fmt.Println(local.Equal(utc), local == utc) // compare instants with Equal, not ==

	// ParseInLocation interprets a time without an offset in a given zone.
	p, _ := time.ParseInLocation("2006-01-02 15:04", "2026-10-16 14:30", ist)
	fmt.Println(p.UTC())

	// Output:
	// 2026-10-16 14:30:00 +0530 IST
	// true false
	// 2026-10-16 09:00:00 +0000 UTC
}

// [end zones]

// [start monotonic]

func Example_measuring() {
	start := time.Now() // carries a monotonic clock reading
	time.Sleep(10 * time.Millisecond)
	elapsed := time.Since(start) // uses the monotonic reading: safe even if the wall clock jumps
	fmt.Println(elapsed >= 10*time.Millisecond)

	// Output:
	// true
}

// [end monotonic]

// [start arithmetic]

func Example_arithmetic() {
	t := time.Date(2026, time.January, 31, 12, 0, 0, 0, time.UTC)

	fmt.Println(t.Add(48 * time.Hour).Format("Jan 2"))
	fmt.Println(t.AddDate(0, 1, 0).Format("Jan 2")) // Feb 31 normalizes to Mar 3

	deadline := t.Add(time.Hour)
	fmt.Println(deadline.Sub(t), deadline.After(t))

	fmt.Println(t.Unix(), time.Unix(0, 0).UTC().Year())

	// Output:
	// Feb 2
	// Mar 3
	// 1h0m0s true
	// 1769860800 1970
}

// [end arithmetic]
