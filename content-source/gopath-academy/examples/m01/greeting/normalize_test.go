// [start internal]

// Same package name as the code under test, so unexported
// identifiers like normalize are visible here.
package greeting

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":         "Gopher",
		"  Ana  ":  "Ana",
		"\tGo\n":   "Go",
		"Dharshan": "Dharshan",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

// [end internal]
