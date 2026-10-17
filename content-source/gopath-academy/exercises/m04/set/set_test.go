package set

import (
	"slices"
	"testing"
)

func TestBasics(t *testing.T) {
	s := New("go", "rust", "go")
	if s.Len() != 2 {
		t.Errorf("Len = %d, want 2", s.Len())
	}
	if !s.Has("go") || s.Has("zig") {
		t.Error("Has gives wrong answers")
	}
	s.Add("zig")
	s.Add("zig")
	if s.Len() != 3 || !s.Has("zig") {
		t.Errorf("after Add: Len = %d", s.Len())
	}
}

func TestEmpty(t *testing.T) {
	s := New[int]()
	if s.Len() != 0 || s.Has(0) {
		t.Error("New() should be empty")
	}
	s.Add(0)
	if !s.Has(0) {
		t.Error("Add on a set made with New() should work")
	}
}

func TestUnion(t *testing.T) {
	a := New(1, 2, 3)
	b := New(3, 4)
	u := a.Union(b)
	if got := Sorted(u); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("Union = %v", got)
	}
	if a.Len() != 3 || b.Len() != 2 {
		t.Error("Union modified its inputs")
	}
	u.Add(99)
	if a.Has(99) {
		t.Error("Union result shares storage with a")
	}
}

func TestSorted(t *testing.T) {
	if got := Sorted(New("pear", "apple", "fig")); !slices.Equal(got, []string{"apple", "fig", "pear"}) {
		t.Errorf("Sorted = %v", got)
	}
}

type point struct{ X, Y int }

func TestStructElements(t *testing.T) {
	// Any comparable type works, including structs.
	s := New(point{1, 2}, point{1, 2}, point{3, 4})
	if s.Len() != 2 || !s.Has(point{3, 4}) {
		t.Errorf("struct set: Len = %d", s.Len())
	}
}
