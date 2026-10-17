package parallel

import (
	"sync"
	"testing"
)

func TestSum(t *testing.T) {
	nums := make([]int, 1001)
	want := 0
	for i := range nums {
		nums[i] = i
		want += i
	}
	for _, workers := range []int{1, 2, 3, 7, 64, 5000} {
		if got := Sum(nums, workers); got != want {
			t.Errorf("Sum(0..1000, %d workers) = %d, want %d", workers, got, want)
		}
	}
	if got := Sum(nil, 4); got != 0 {
		t.Errorf("Sum(nil) = %d", got)
	}
	if got := Sum([]int{5}, 3); got != 5 {
		t.Errorf("Sum([5], 3) = %d", got)
	}
}

func TestTallyConcurrent(t *testing.T) {
	var tally Tally // zero value must work
	var wg sync.WaitGroup
	for i := range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tally.Add("go")
			if i%2 == 0 {
				tally.Add("rust")
			}
			_ = tally.Count("go") // concurrent reads too
		}()
	}
	wg.Wait()
	if got := tally.Count("go"); got != 200 {
		t.Errorf("go = %d, want 200", got)
	}
	if got := tally.Count("rust"); got != 100 {
		t.Errorf("rust = %d, want 100", got)
	}
}

func TestSnapshotIsACopy(t *testing.T) {
	var tally Tally
	if snap := tally.Snapshot(); snap == nil || len(snap) != 0 {
		t.Errorf("Snapshot of empty Tally = %v, want an empty non-nil map", snap)
	}
	tally.Add("x")
	snap := tally.Snapshot()
	snap["x"] = 100
	if tally.Count("x") != 1 {
		t.Error("changing the snapshot changed the Tally")
	}
}
