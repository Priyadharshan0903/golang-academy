package seq

import (
	"iter"
	"slices"
	"testing"
)

// naturals is an infinite iterator: 0, 1, 2, ...
// It records how many values were requested so tests can check
// that Take stops pulling.
func naturals(pulled *int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; ; i++ {
			*pulled++
			if !yield(i) {
				return
			}
		}
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		start, end, step int
		want             []int
	}{
		{0, 10, 3, []int{0, 3, 6, 9}},
		{1, 4, 1, []int{1, 2, 3}},
		{5, 5, 1, nil},
		{7, 3, 1, nil},
	}
	for _, tt := range tests {
		got := slices.Collect(Range(tt.start, tt.end, tt.step))
		if !slices.Equal(got, tt.want) {
			t.Errorf("Range(%d, %d, %d) = %v, want %v", tt.start, tt.end, tt.step, got, tt.want)
		}
	}
}

func TestRangeEarlyBreak(t *testing.T) {
	var got []int
	for v := range Range(0, 100, 1) {
		if v == 3 {
			break
		}
		got = append(got, v)
	}
	if !slices.Equal(got, []int{0, 1, 2}) {
		t.Errorf("got %v, want [0 1 2]", got)
	}
}

func TestMap(t *testing.T) {
	sq := Map(Range(1, 5, 1), func(v int) int { return v * v })
	if got := slices.Collect(sq); !slices.Equal(got, []int{1, 4, 9, 16}) {
		t.Errorf("Map squares = %v", got)
	}
}

func TestTake(t *testing.T) {
	pulled := 0
	got := slices.Collect(Take(naturals(&pulled), 3))
	if !slices.Equal(got, []int{0, 1, 2}) {
		t.Errorf("Take(naturals, 3) = %v", got)
	}
	if pulled != 3 {
		t.Errorf("Take pulled %d values from the source, want exactly 3", pulled)
	}
	if got := slices.Collect(Take(Range(0, 2, 1), 5)); !slices.Equal(got, []int{0, 1}) {
		t.Errorf("Take more than available = %v", got)
	}
	if got := slices.Collect(Take(Range(0, 2, 1), 0)); len(got) != 0 {
		t.Errorf("Take 0 = %v", got)
	}
}

func TestComposition(t *testing.T) {
	pulled := 0
	pipeline := Take(Map(naturals(&pulled), func(v int) int { return v * 10 }), 4)
	if got := slices.Collect(pipeline); !slices.Equal(got, []int{0, 10, 20, 30}) {
		t.Errorf("pipeline = %v", got)
	}
}
