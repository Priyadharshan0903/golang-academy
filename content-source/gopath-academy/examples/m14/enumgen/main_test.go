package main

import (
	"bytes"
	"os"
	"testing"
)

// [start uptodate]

// TestGeneratedCodeIsUpToDate fails if someone changed the constants in
// package color but forgot to run go generate.
func TestGeneratedCodeIsUpToDate(t *testing.T) {
	want, err := generate("../color", "Color")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile("../color/color_enum.go")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("color_enum.go is stale: run go generate ./examples/m14/...")
	}
}

// [end uptodate]

func TestErrors(t *testing.T) {
	if _, err := generate("../color", "Nope"); err == nil {
		t.Error("expected an error for a missing type")
	}
	if _, err := generate(t.TempDir(), "X"); err == nil {
		t.Error("expected an error for an empty directory")
	}
}
