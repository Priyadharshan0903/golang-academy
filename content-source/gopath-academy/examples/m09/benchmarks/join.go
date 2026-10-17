package benchmarks

import (
	"strconv"
	"strings"
)

// [start code]

// JoinPlus builds the string with repeated concatenation.
func JoinPlus(nums []int) string {
	s := ""
	for i, n := range nums {
		if i > 0 {
			s += ","
		}
		s += strconv.Itoa(n)
	}
	return s
}

// JoinBuilder uses a strings.Builder with a size estimate.
func JoinBuilder(nums []int) string {
	var b strings.Builder
	b.Grow(len(nums) * 6) // estimate: up to 5 digits plus a comma
	for i, n := range nums {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(n))
	}
	return b.String()
}

// JoinAppend writes digits straight into a byte slice.
func JoinAppend(nums []int) string {
	// An earlier version estimated 4 bytes per number. The allocation
	// test below caught that: for 4-digit numbers plus commas, the
	// buffer had to grow once, costing an extra allocation and a copy.
	buf := make([]byte, 0, len(nums)*6)
	for i, n := range nums {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = strconv.AppendInt(buf, int64(n), 10)
	}
	return string(buf)
}

// [end code]
