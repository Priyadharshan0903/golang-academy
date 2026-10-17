// Package count counts word frequencies in text.
package count

import (
	"bufio"
	"cmp"
	"io"
	"maps"
	"slices"
	"strings"
	"unicode"
)

// Entry is one word and how often it appeared.
type Entry struct {
	Word  string
	Count int
}

// Words reads all of r and returns how often each word appears.
//
// A word is a run of letters, combining marks, digits and apostrophes.
// Marks matter: vowel signs in scripts like Tamil or Hindi are marks,
// not letters, and treating them as separators would split words.
// Words are
// lowercased, and leading or trailing apostrophes are removed, so
// "Don't" counts as "don't" and 'quoted' counts as "quoted".
func Words(r io.Reader) (map[string]int, error) {
	counts := make(map[string]int)
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024) // allow long lines
	for sc.Scan() {
		for _, w := range split(sc.Text()) {
			counts[w]++
		}
	}
	return counts, sc.Err()
}

// split breaks one line into normalized words.
func split(line string) []string {
	fields := strings.FieldsFunc(line, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsMark(r) && !unicode.IsDigit(r) && r != '\''
	})
	words := fields[:0] // filter in place, reusing fields' array
	for _, f := range fields {
		f = strings.Trim(f, "'")
		if f != "" {
			words = append(words, strings.ToLower(f))
		}
	}
	return words
}

// Top returns the n most frequent words, most frequent first. Words
// with equal counts are ordered alphabetically, so the result is
// deterministic even though map iteration order is not.
// If n <= 0 or n exceeds the number of words, all words are returned.
func Top(counts map[string]int, n int) []Entry {
	entries := make([]Entry, 0, len(counts))
	for _, w := range slices.Sorted(maps.Keys(counts)) {
		entries = append(entries, Entry{Word: w, Count: counts[w]})
	}
	slices.SortStableFunc(entries, func(a, b Entry) int {
		return cmp.Compare(b.Count, a.Count) // descending by count
	})
	if n > 0 && n < len(entries) {
		entries = entries[:n]
	}
	return entries
}
