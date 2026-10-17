package loganalyzer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// [start naive]

// AnalyzeNaive is written the way code often starts: clear, but with
// no thought for cost.
func AnalyzeNaive(lines []string) Stats {
	s := Stats{ByStatus: map[int]int{}, ByPath: map[string]int{}}
	for _, line := range lines {
		// Compiling a regular expression on every line is expensive.
		re := regexp.MustCompile(`^(\S+) (\S+) (\S+) (\d{3}) (\d+)ms$`)
		m := re.FindStringSubmatch(line)
		if m == nil {
			s.Malformed++
			continue
		}
		status, _ := strconv.Atoi(m[4])
		ms, _ := strconv.Atoi(m[5])

		// Building a key with Sprintf allocates a new string per line.
		key := fmt.Sprintf("%s", strings.ToLower(m[3]))

		s.Requests++
		s.ByStatus[status]++
		s.ByPath[key]++
		s.TotalMs += ms
	}
	return s
}

// [end naive]
