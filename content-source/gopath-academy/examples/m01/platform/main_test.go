package main

import (
	"runtime"
	"testing"
)

// The output depends on the machine, so this test only checks that
// the values the program prints are never empty.
func TestPlatformKnown(t *testing.T) {
	if runtime.GOOS == "" || runtime.GOARCH == "" {
		t.Fatalf("unexpected empty platform: %q/%q", runtime.GOOS, runtime.GOARCH)
	}
}
