package count

import (
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestWords(t *testing.T) {
	in := "The cat, the HAT; and 'the' bat!\nDon't stop. don't 42 times 42"
	got, err := Words(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{
		"the": 3, "cat": 1, "hat": 1, "and": 1, "bat": 1,
		"don't": 2, "stop": 1, "42": 2, "times": 1,
	}
	if !maps.Equal(got, want) {
		t.Errorf("Words() =\n%v\nwant\n%v", got, want)
	}
}

func TestWordsUnicode(t *testing.T) {
	got, err := Words(strings.NewReader("Café café CAFÉ, வணக்கம் வணக்கம்"))
	if err != nil {
		t.Fatal(err)
	}
	if got["café"] != 3 {
		t.Errorf("café counted %d times, want 3", got["café"])
	}
	if len(got) != 2 {
		t.Errorf("got %d distinct words, want 2: %v", len(got), got)
	}
}

func TestWordsEmpty(t *testing.T) {
	got, err := Words(strings.NewReader(""))
	if err != nil || len(got) != 0 {
		t.Errorf("Words(\"\") = %v, %v", got, err)
	}
}

func TestTop(t *testing.T) {
	counts := map[string]int{"b": 2, "a": 2, "c": 5, "d": 1}
	got := Top(counts, 3)
	want := []Entry{{"c", 5}, {"a", 2}, {"b", 2}}
	if !slices.Equal(got, want) {
		t.Errorf("Top(3) = %v, want %v", got, want)
	}
	if all := Top(counts, 0); len(all) != 4 {
		t.Errorf("Top(0) returned %d entries, want all 4", len(all))
	}
	if all := Top(counts, 99); len(all) != 4 {
		t.Errorf("Top(99) returned %d entries, want 4", len(all))
	}
}

func TestSampleFile(t *testing.T) {
	f, err := os.Open("testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	counts, err := Words(f)
	if err != nil {
		t.Fatal(err)
	}
	top := Top(counts, 2)
	want := []Entry{{"go", 4}, {"is", 3}}
	if !slices.Equal(top, want) {
		t.Errorf("Top(2) of sample = %v, want %v", top, want)
	}
}
