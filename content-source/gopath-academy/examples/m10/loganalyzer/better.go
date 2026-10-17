package loganalyzer

import (
	"regexp"
	"strconv"
	"strings"
)

// [start better]

// lineRE is compiled once, when the package is initialized.
var lineRE = regexp.MustCompile(`^(\S+) (\S+) (\S+) (\d{3}) (\d+)ms$`)

// AnalyzeBetter applies the first fix the profile points at: compile
// the regular expression once. Everything else is unchanged.
func AnalyzeBetter(lines []string) Stats {
	s := Stats{ByStatus: map[int]int{}, ByPath: map[string]int{}}
	for _, line := range lines {
		m := lineRE.FindStringSubmatch(line)
		if m == nil {
			s.Malformed++
			continue
		}
		status, _ := strconv.Atoi(m[4])
		ms, _ := strconv.Atoi(m[5])
		s.Requests++
		s.ByStatus[status]++
		s.ByPath[strings.ToLower(m[3])]++
		s.TotalMs += ms
	}
	return s
}

// [end better]
