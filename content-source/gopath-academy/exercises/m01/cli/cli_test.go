package cli

import (
	"errors"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse(nil) returned error: %v", err)
	}
	want := Config{Name: "Gopher", Shout: false, Times: 1}
	if cfg != want {
		t.Errorf("Parse(nil) = %+v, want %+v", cfg, want)
	}
}

func TestParseFlags(t *testing.T) {
	cfg, err := Parse([]string{"-name", "Ana", "-shout", "-times=3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Config{Name: "Ana", Shout: true, Times: 3}
	if cfg != want {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestParseInvalidTimes(t *testing.T) {
	for _, v := range []string{"0", "-2"} {
		_, err := Parse([]string{"-times", v})
		if !errors.Is(err, ErrInvalidTimes) {
			t.Errorf("-times %s: got error %v, want ErrInvalidTimes", v, err)
		}
	}
}

func TestParseUnknownFlag(t *testing.T) {
	if _, err := Parse([]string{"-nope"}); err == nil {
		t.Error("expected an error for an unknown flag")
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"once", Config{Name: "Ana", Times: 1}, "Hello, Ana!"},
		{"twice", Config{Name: "Ana", Times: 2}, "Hello, Ana!\nHello, Ana!"},
		{"shout", Config{Name: "Ana", Times: 1, Shout: true}, "HELLO, ANA!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Render(tt.cfg); got != tt.want {
				t.Errorf("Render(%+v) = %q, want %q", tt.cfg, got, tt.want)
			}
		})
	}
}
