package loganalyzer

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"testing"
)

// sampleLines generates a realistic, reproducible log.
func sampleLines(n int) []string {
	r := rand.New(rand.NewPCG(1, 2))
	paths := []string{"/api/users", "/api/orders", "/health", "/api/users/42", "/static/app.js", "/API/Users"}
	codes := []int{200, 200, 200, 201, 304, 404, 500}
	lines := make([]string, n)
	for i := range lines {
		if i%97 == 0 {
			lines[i] = "garbage line"
			continue
		}
		lines[i] = fmt.Sprintf("2026-10-16T09:%02d:%02dZ GET %s %d %dms",
			i/60%60, i%60, paths[r.IntN(len(paths))], codes[r.IntN(len(codes))], r.IntN(500))
	}
	return lines
}

func equal(a, b Stats) bool {
	return a.Requests == b.Requests && a.TotalMs == b.TotalMs && a.Malformed == b.Malformed &&
		maps.Equal(a.ByStatus, b.ByStatus) && maps.Equal(a.ByPath, b.ByPath)
}

// [start equivalence]

func TestImplementationsAgree(t *testing.T) {
	lines := append(sampleLines(5000),
		"", "x GET /a 200", "x GET /a 2000 5ms", "x GET /a 200 -5ms", "x GET /a 200 5ms extra")
	want := AnalyzeNaive(lines)
	if want.Requests == 0 || want.Malformed == 0 {
		t.Fatalf("sample data looks wrong: %+v", want)
	}
	for name, f := range map[string]func([]string) Stats{"better": AnalyzeBetter, "fast": AnalyzeFast} {
		if got := f(lines); !equal(got, want) {
			t.Errorf("%s disagrees with naive:\n got %+v\nwant %+v", name, got, want)
		}
	}
}

// [end equivalence]

// [start bench]

func BenchmarkAnalyze(b *testing.B) {
	lines := sampleLines(10_000)
	impls := []struct {
		name string
		f    func([]string) Stats
	}{
		{"naive", AnalyzeNaive},
		{"better", AnalyzeBetter},
		{"fast", AnalyzeFast},
	}
	for _, impl := range impls {
		b.Run("impl="+impl.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				impl.f(lines)
			}
		})
	}
}

// [end bench]

// Single-implementation benchmarks, convenient for profiling one at a time.
func BenchmarkNaive(b *testing.B) {
	lines := sampleLines(10_000)
	for b.Loop() {
		AnalyzeNaive(lines)
	}
}

func BenchmarkFast(b *testing.B) {
	lines := sampleLines(10_000)
	for b.Loop() {
		AnalyzeFast(lines)
	}
}
