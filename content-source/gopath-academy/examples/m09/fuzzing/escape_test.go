package fuzzing

import "testing"

// [start fuzz]

// FuzzRoundTrip checks a property that must hold for EVERY input:
// unquoting a quoted string gives back the original.
//
//	go test -fuzz=FuzzRoundTrip -fuzztime=30s ./examples/m09/fuzzing
//
// Without -fuzz, plain `go test` runs only the seed corpus below (and
// any saved failures in testdata/fuzz), so it is fast enough for CI.
func FuzzRoundTrip(f *testing.F) {
	for _, seed := range []string{"", "plain", `with "quotes"`, `back\slash`, "unicode é 世界"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		q := Quote(s)
		got, err := Unquote(q)
		if err != nil {
			t.Fatalf("Unquote(Quote(%q)) error: %v", s, err)
		}
		if got != s {
			t.Fatalf("round trip changed %q into %q", s, got)
		}
	})
}

// FuzzUnquoteNoPanic only checks that arbitrary input never panics.
func FuzzUnquoteNoPanic(f *testing.F) {
	f.Add(`"ok"`)
	f.Add(`"\`)
	f.Fuzz(func(t *testing.T, q string) {
		Unquote(q) // errors are fine; panics are not
	})
}

// [end fuzz]
