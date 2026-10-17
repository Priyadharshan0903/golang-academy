package wordwrap

import (
	"slices"
	"testing"
)

// This is YOUR test file. It must pass against the correct Wrap, and
// fail against every buggy version. Check with:
//
//	cd exercises
//	go test ./m09/wordwrap                              # must PASS
//	for b in bug_overflow bug_longword bug_empty bug_lastline; do
//	    go test -tags $b ./m09/wordwrap > /dev/null && echo "$b NOT caught"
//	done                                                 # must print nothing
//
// Turn this into a table-driven test with enough cases to catch them all.

func TestWrap(t *testing.T) {
	got := Wrap("the quick brown fox", 10)
	want := []string{"the quick", "brown fox"}
	if !slices.Equal(got, want) {
		t.Errorf("Wrap = %q, want %q", got, want)
	}
	// TODO: add cases until every bug_* version fails.
}
