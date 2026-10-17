// Package version holds build metadata that is injected at build time.
//
// Build with:
//
//	go build -ldflags "-X github.com/Priyadharshan0903/gopath-academy/projects/greet/internal/version.Version=v0.1.0" ./cmd/greet
//
// The Makefile does this for you.
package version

import "fmt"

// These must be variables (not constants) for -ldflags -X to work.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns a single line describing this build.
func String() string {
	return fmt.Sprintf("greet %s (commit %s, built %s)", Version, Commit, Date)
}
