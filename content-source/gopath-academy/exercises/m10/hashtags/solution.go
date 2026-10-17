//go:build solution

// Package hashtags is an optimization exercise.
package hashtags

import (
	"cmp"
	"slices"
	"strings"
)

// TopTags returns the n most common hashtags in posts.
func TopTags(posts []string, n int) []string {
	counts := make(map[string]int, 64)
	for _, post := range posts {
		for i := 0; i < len(post); i++ {
			if post[i] != '#' || (i > 0 && !isSpace(post[i-1])) {
				continue // a tag must start a word
			}
			j := i + 1
			lower := true
			for j < len(post) && isTagByte(post[j]) {
				if 'A' <= post[j] && post[j] <= 'Z' {
					lower = false
				}
				j++
			}
			if j == i+1 {
				continue
			}
			tag := post[i+1 : j] // a substring: no allocation
			if !lower {
				tag = strings.ToLower(tag) // allocate only when needed
			}
			counts[tag]++
			i = j - 1 // continue after the tag
		}
	}

	type pair struct {
		tag   string
		count int
	}
	pairs := make([]pair, 0, len(counts))
	for t, c := range counts {
		pairs = append(pairs, pair{t, c})
	}
	slices.SortFunc(pairs, func(a, b pair) int {
		return cmp.Or(cmp.Compare(b.count, a.count), strings.Compare(a.tag, b.tag))
	})

	out := make([]string, 0, min(n, len(pairs)))
	for _, p := range pairs[:min(n, len(pairs))] {
		out = append(out, p.tag)
	}
	return out
}

func isTagByte(c byte) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}
