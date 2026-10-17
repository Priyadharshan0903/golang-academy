//go:build !solution

// Package hashtags is an optimization exercise. The starter code below
// is CORRECT but allocates far too much. Make it pass the allocation
// budget test without breaking the correctness tests.
//
//	cd exercises
//	go test ./m10/hashtags                       # correctness + budget
//	go test -bench . -benchmem ./m10/hashtags    # measure
//	go test -bench . -memprofile mem.out ./m10/hashtags && go tool pprof -top -sample_index=alloc_space mem.out
package hashtags

import (
	"fmt"
	"sort"
	"strings"
)

// TopTags returns the n most common hashtags in posts, most common
// first, ties broken alphabetically. A hashtag is a word (words are
// separated by ASCII whitespace) that starts with '#' followed by one or
// more ASCII letters, digits or underscores; the tag ends at the first
// other character. Tags are case-insensitive and returned in lowercase.
func TopTags(posts []string, n int) []string {
	counts := map[string]int{}
	for _, post := range posts {
		for _, word := range strings.Fields(post) { // allocates a slice per post
			if !strings.HasPrefix(word, "#") {
				continue
			}
			tag := ""
			for _, r := range word[1:] {
				if r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
					tag += string(r) // allocates on every character
				} else {
					break
				}
			}
			if tag != "" {
				counts[strings.ToLower(tag)]++ // allocates even when already lowercase
			}
		}
	}

	type pair struct {
		tag   string
		count int
	}
	var pairs []pair
	for t, c := range counts {
		pairs = append(pairs, pair{t, c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		// Building sort keys with Sprintf allocates on every comparison.
		return fmt.Sprintf("%09d", 1e9-pairs[i].count)+pairs[i].tag < fmt.Sprintf("%09d", 1e9-pairs[j].count)+pairs[j].tag
	})

	var out []string
	for i := 0; i < len(pairs) && i < n; i++ {
		out = append(out, pairs[i].tag)
	}
	return out
}
