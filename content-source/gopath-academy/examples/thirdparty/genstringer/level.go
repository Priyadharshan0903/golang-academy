// Package genstringer shows a code generator run with `go tool`
// (lesson 14.3). The stringer tool is declared in this module's go.mod
// with a tool directive (Go 1.24+), so everyone gets the same version.
package genstringer

//go:generate go tool stringer -type=Level -trimprefix=Level

// Level is a log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)
