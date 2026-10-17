package ints

import (
	"slices"
	"testing"
)

func TestChunk(t *testing.T) {
	tests := []struct {
		in   []int
		size int
		want [][]int
	}{
		{nil, 3, nil},
		{[]int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{[]int{1, 2, 3}, 3, [][]int{{1, 2, 3}}},
		{[]int{1, 2}, 5, [][]int{{1, 2}}},
		{[]int{1, 2, 3}, 1, [][]int{{1}, {2}, {3}}},
	}
	for _, tt := range tests {
		got := Chunk(tt.in, tt.size)
		if !slices.EqualFunc(got, tt.want, slices.Equal[[]int]) {
			t.Errorf("Chunk(%v, %d) = %v, want %v", tt.in, tt.size, got, tt.want)
		}
	}
}

func TestChunkDoesNotAlias(t *testing.T) {
	s := make([]int, 4, 10)
	copy(s, []int{1, 2, 3, 4})
	chunks := Chunk(s, 2)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	_ = append(chunks[0], 99) // must not overwrite chunks[1] or s
	if s[2] != 3 || chunks[1][0] != 3 {
		t.Errorf("appending to a chunk changed other data: s=%v chunks=%v", s, chunks)
	}
}

func TestChunkPanicsOnBadSize(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Chunk with size 0 should panic")
		}
	}()
	Chunk([]int{1}, 0)
}

func TestRemoveAt(t *testing.T) {
	s := []int{10, 20, 30, 40}
	got := RemoveAt(s, 1)
	if !slices.Equal(got, []int{10, 30, 40}) {
		t.Errorf("RemoveAt(..., 1) = %v, want [10 30 40]", got)
	}
	if !slices.Equal(s, []int{10, 20, 30, 40}) {
		t.Errorf("RemoveAt modified its input: %v", s)
	}
	if got := RemoveAt([]int{7}, 0); len(got) != 0 {
		t.Errorf("RemoveAt([7], 0) = %v, want []", got)
	}
	if got := RemoveAt(s, 3); !slices.Equal(got, []int{10, 20, 30}) {
		t.Errorf("RemoveAt(..., last) = %v", got)
	}
}

func TestDedupe(t *testing.T) {
	in := []int{3, 1, 3, 2, 1, 3}
	got := Dedupe(in)
	if !slices.Equal(got, []int{3, 1, 2}) {
		t.Errorf("Dedupe(%v) = %v, want [3 1 2]", in, got)
	}
	if len(Dedupe(nil)) != 0 {
		t.Error("Dedupe(nil) should be empty")
	}
	got[0] = 100
	if in[0] != 3 {
		t.Error("Dedupe result shares memory with its input")
	}
}
