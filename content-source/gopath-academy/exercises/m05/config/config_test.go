package config

import (
	"errors"
	"maps"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	k, v, err := ParseLine("  port =  8080 ")
	if err != nil || k != "port" || v != "8080" {
		t.Errorf("ParseLine = %q, %q, %v", k, v, err)
	}
	k, v, err = ParseLine("empty=")
	if err != nil || k != "empty" || v != "" {
		t.Errorf("empty value: %q, %q, %v", k, v, err)
	}
	k, v, err = ParseLine("url = http://x?a=b")
	if err != nil || v != "http://x?a=b" {
		t.Errorf("only the first = separates: got %q, %q, %v", k, v, err)
	}
	for _, bad := range []string{"no equals sign", " = value"} {
		_, _, err := ParseLine(bad)
		if !errors.Is(err, ErrSyntax) {
			t.Errorf("ParseLine(%q) error = %v, want one wrapping ErrSyntax", bad, err)
			continue
		}
		if !strings.Contains(err.Error(), strings.TrimSpace(bad)) {
			t.Errorf("error %q should mention the bad line", err)
		}
	}
}

func TestLoad(t *testing.T) {
	in := "# server settings\nhost = localhost\n\nport = 8080\n  # indented comment\n"
	got, err := Load(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]string{"host": "localhost", "port": "8080"}
	if !maps.Equal(got, want) {
		t.Errorf("Load = %v, want %v", got, want)
	}
}

func TestLoadSyntaxError(t *testing.T) {
	_, err := Load(strings.NewReader("a = 1\n\nthis is wrong\nb = 2\n"))

	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("error %v is not a *ParseError", err)
	}
	if pe.Line != 3 {
		t.Errorf("Line = %d, want 3", pe.Line)
	}
	if !errors.Is(err, ErrSyntax) {
		t.Error("errors.Is(err, ErrSyntax) should be true through ParseError.Unwrap")
	}
	if !strings.HasPrefix(err.Error(), "line 3: ") {
		t.Errorf("Error() = %q, want it to start with \"line 3: \"", err.Error())
	}
}

func TestLoadDuplicate(t *testing.T) {
	_, err := Load(strings.NewReader("a = 1\nb = 2\na = 3\n"))
	var pe *ParseError
	if !errors.As(err, &pe) || pe.Line != 3 {
		t.Fatalf("got %v, want a *ParseError on line 3", err)
	}
	if !errors.Is(err, ErrDuplicateKey) || errors.Is(err, ErrSyntax) {
		t.Errorf("duplicate key error should wrap ErrDuplicateKey only: %v", err)
	}
}
