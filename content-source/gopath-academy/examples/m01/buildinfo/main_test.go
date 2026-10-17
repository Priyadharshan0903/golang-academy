package main

import (
	"fmt"
	"testing"
)

func Example_versionString() {
	// Without -ldflags, the defaults from the source are used.
	fmt.Println(versionString())
	// Output: version=dev commit=none
}

func TestModuleVersionNotEmpty(t *testing.T) {
	if moduleVersion() == "" {
		t.Fatal("moduleVersion returned an empty string")
	}
}
