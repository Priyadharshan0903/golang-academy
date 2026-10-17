//go:build solution

package wordwrap

import (
	"slices"
	"strings"
	"testing"
)

// TestWrapReference is the reference solution: a table that catches
// every bug, plus a property check that no line exceeds the width.
func TestWrapReference(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  []string
	}{
		{"fits exactly", "ab cd", 5, []string{"ab cd"}},                            // catches nothing alone...
		{"one over", "ab cd", 4, []string{"ab", "cd"}},                             // ...this catches bug_overflow
		{"long word", "a superlongword b", 5, []string{"a", "superlongword", "b"}}, // bug_longword
		{"empty", "", 10, nil},                                                     // bug_empty
		{"spaces only", "   \t\n", 10, nil},                                        // bug_empty
		{"three lines", "aa bb cc", 2, []string{"aa", "bb", "cc"}},                 // bug_lastline
		{"collapse spaces", "a    b\n\tc", 80, []string{"a b c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.text, tt.width)
			if !slices.Equal(got, tt.want) || (tt.want == nil && got != nil) {
				t.Errorf("Wrap(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.want)
			}
		})
	}
}

func TestWrapProperties(t *testing.T) {
	text := strings.Repeat("lorem ipsum dolor sit amet consectetur ", 20)
	for width := 5; width <= 40; width++ {
		lines := Wrap(text, width)
		for _, l := range lines {
			if len(l) > width && strings.Contains(l, " ") {
				t.Fatalf("width %d: line %q is too long", width, l)
			}
		}
		if got := strings.Join(lines, " "); got != strings.Join(strings.Fields(text), " ") {
			t.Fatalf("width %d: words were lost or reordered", width)
		}
	}
}
