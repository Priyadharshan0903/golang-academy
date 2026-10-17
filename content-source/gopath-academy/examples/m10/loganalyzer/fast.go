package loganalyzer

import (
	"strconv"
	"strings"
)

// [start fast]

// AnalyzeFast replaces the regular expression with direct string
// scanning, avoids allocating keys that are already lowercase, and
// sizes the maps up front.
func AnalyzeFast(lines []string) Stats {
	s := Stats{ByStatus: make(map[int]int, 8), ByPath: make(map[string]int, 64)}
	for _, line := range lines {
		path, status, ms, ok := parseLine(line)
		if !ok {
			s.Malformed++
			continue
		}
		s.Requests++
		s.ByStatus[status]++
		s.ByPath[lowerIfNeeded(path)]++
		s.TotalMs += ms
	}
	return s
}

// parseLine splits "time method path status NNms" without allocating.
func parseLine(line string) (path string, status, ms int, ok bool) {
	_, rest, ok1 := strings.Cut(line, " ") // time
	_, rest, ok2 := strings.Cut(rest, " ") // method
	path, rest, ok3 := strings.Cut(rest, " ")
	code, dur, ok4 := strings.Cut(rest, " ")
	if !(ok1 && ok2 && ok3 && ok4) || len(code) != 3 || !strings.HasSuffix(dur, "ms") || strings.Contains(dur, " ") {
		return "", 0, 0, false
	}
	status, err1 := strconv.Atoi(code)
	ms, err2 := strconv.Atoi(dur[:len(dur)-2])
	if err1 != nil || err2 != nil || ms < 0 || path == "" {
		return "", 0, 0, false
	}
	return path, status, ms, true
}

// lowerIfNeeded returns s unchanged (no allocation) when it has no
// uppercase ASCII letters, which is the common case for URL paths.
func lowerIfNeeded(s string) string {
	for i := 0; i < len(s); i++ {
		if c := s[i]; 'A' <= c && c <= 'Z' || c >= 0x80 {
			return strings.ToLower(s)
		}
	}
	return s
}

// [end fast]
