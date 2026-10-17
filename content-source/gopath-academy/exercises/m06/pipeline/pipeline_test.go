package pipeline

import (
	"runtime"
	"slices"
	"testing"
	"time"
)

func collect(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	return out
}

func TestGenerate(t *testing.T) {
	if got := collect(Generate(3, 1, 2)); !slices.Equal(got, []int{3, 1, 2}) {
		t.Errorf("Generate = %v, want [3 1 2] in order", got)
	}
	if got := collect(Generate()); len(got) != 0 {
		t.Errorf("Generate() = %v, want nothing", got)
	}
}

func TestSquare(t *testing.T) {
	if got := collect(Square(Generate(1, 2, 3, 4))); !slices.Equal(got, []int{1, 4, 9, 16}) {
		t.Errorf("Square = %v", got)
	}
}

func TestMerge(t *testing.T) {
	got := collect(Merge(Generate(1, 2), Generate(10, 20, 30), Generate()))
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2, 10, 20, 30}) {
		t.Errorf("Merge = %v", got)
	}
	if got := collect(Merge()); len(got) != 0 {
		t.Errorf("Merge() = %v, want nothing and a closed channel", got)
	}
}

func TestFullPipeline(t *testing.T) {
	got := collect(Merge(Square(Generate(1, 2)), Square(Generate(3))))
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 4, 9}) {
		t.Errorf("pipeline = %v", got)
	}
}

func TestNoGoroutineLeaks(t *testing.T) {
	before := runtime.NumGoroutine()
	for range 50 {
		collect(Merge(Square(Generate(1, 2, 3)), Generate(4)))
	}
	// Give finished goroutines a moment to exit.
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Errorf("%d goroutines still running after the pipelines finished (started with %d)", n, before)
	}
}
