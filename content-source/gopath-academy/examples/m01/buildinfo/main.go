// Command buildinfo shows two ways a Go binary can know its own version.
package main

import (
	"fmt"
	"runtime/debug"
)

// [start ldflags]

// These are overwritten at build time with:
//
//	go build -ldflags "-X main.version=1.2.0 -X main.commit=abc123" .
//
// -X only works on package-level string variables (not constants).
var (
	version = "dev"
	commit  = "none"
)

func versionString() string {
	return fmt.Sprintf("version=%s commit=%s", version, commit)
}

// [end ldflags]

// [start buildinfo]

// moduleVersion reads metadata the Go toolchain embeds in every binary.
// After `go install example.com/tool@v1.4.0`, Main.Version is "v1.4.0".
// For a local `go build` it is "(devel)" or a pseudo-version
// derived from your VCS state.
func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
}

// [end buildinfo]

func main() {
	fmt.Println(versionString())
	fmt.Println("module version:", moduleVersion())
}
