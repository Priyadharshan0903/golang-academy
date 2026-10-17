package hashtags

import (
	"fmt"
	"slices"
	"testing"
)

func TestTopTags(t *testing.T) {
	posts := []string{
		"loving #golang and #kubernetes today",
		"#GoLang is fast #go_1_27",
		"#golang, #kubernetes! #k8s",
		"no tags here, just a # sign and email@example.com",
		"#k8s #k8s #Kubernetes",
		"",
	}
	got := TopTags(posts, 3)
	want := []string{"golang", "k8s", "kubernetes"} // all three appear 3 times: ties are alphabetical
	if !slices.Equal(got, want) {
		t.Errorf("TopTags = %q, want %q", got, want)
	}
	if all := TopTags(posts, 100); !slices.Equal(all, []string{"golang", "k8s", "kubernetes", "go_1_27"}) {
		t.Errorf("all tags = %q", all)
	}
	if got := TopTags(nil, 5); len(got) != 0 {
		t.Errorf("TopTags(nil) = %q", got)
	}
}

func TestTiesAreAlphabetical(t *testing.T) {
	got := TopTags([]string{"#b #a #c", "#c #a #b"}, 3)
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("got %q", got)
	}
}

func posts(n int) []string {
	tags := []string{"#golang", "#k8s", "#rust", "#Kubernetes", "#devops", "#sre", "#ai", "#chennai"}
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("post %d about %s and %s with some ordinary words in between", i, tags[i%len(tags)], tags[(i*3)%len(tags)])
	}
	return out
}

// TestAllocationBudget fails for the starter code. A good solution
// allocates roughly once per distinct tag plus a small constant, not
// once per post, word or character.
func TestAllocationBudget(t *testing.T) {
	in := posts(1000)
	allocs := testing.AllocsPerRun(20, func() { TopTags(in, 5) })
	const budget = 400
	if allocs > budget {
		t.Errorf("TopTags made %.0f allocations for 1000 posts, budget is %d", allocs, budget)
	}
}

func BenchmarkTopTags(b *testing.B) {
	in := posts(1000)
	b.ReportAllocs()
	for b.Loop() {
		TopTags(in, 5)
	}
}
