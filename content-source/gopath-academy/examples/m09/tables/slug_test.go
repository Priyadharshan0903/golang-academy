package tables

import "testing"

// [start table]

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string // describes the case; becomes the subtest name
		input string
		want  string
	}{
		{"simple", "Hello World", "hello-world"},
		{"punctuation", "Go 1.27: What's New?", "go-1-27-what-s-new"},
		{"repeated separators", "a  --  b", "a-b"},
		{"leading and trailing", "  --Edge Cases--  ", "edge-cases"},
		{"unicode letters", "Café Olé", "café-olé"},
		{"empty", "", ""},
		{"only symbols", "!!!", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// [end table]

// [start parallel]

func TestSlugifyParallel(t *testing.T) {
	tests := map[string]struct{ input, want string }{
		"spaces": {"a b c", "a-b-c"},
		"digits": {"Top 10", "top-10"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel() // run this subtest alongside the others
			if got := Slugify(tt.input); got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// [end parallel]
