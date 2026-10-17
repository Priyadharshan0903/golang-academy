package greeting

import "testing"

func TestGreeting(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain name", "Ana", "Hello, Ana!"},
		{"empty falls back", "", "Hello, Gopher!"},
		{"only spaces falls back", "   ", "Hello, Gopher!"},
		{"trims spaces", "  Dharshan ", "Hello, Dharshan!"},
		{"keeps inner spaces", "Rob Pike", "Hello, Rob Pike!"},
		{"unicode", "Gö", "Hello, Gö!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Greeting(tt.in); got != tt.want {
				t.Errorf("Greeting(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
