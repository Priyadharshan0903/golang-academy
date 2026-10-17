// Package loganalyzer is the running example for Module 10: the same
// analysis written naively and then optimized step by step, guided by
// benchmarks and profiles.
package loganalyzer

// Stats summarizes access log lines of the form
//
//	2026-10-16T09:30:00Z GET /api/users 200 12ms
type Stats struct {
	Requests  int
	ByStatus  map[int]int
	ByPath    map[string]int
	TotalMs   int
	Malformed int
}
