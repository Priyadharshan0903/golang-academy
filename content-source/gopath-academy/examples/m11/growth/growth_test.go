package growth

import (
	"runtime"
	"testing"
)

// [start test]

func TestCapSteps(t *testing.T) {
	t.Logf("heap slice:  %v", CapSteps(300))
	t.Logf("local slice: %v", CapStepsLocal(300))
	for _, steps := range [][]int{CapSteps(300), CapStepsLocal(300)} {
		if last := steps[len(steps)-1]; last < 300 {
			t.Fatalf("final capacity %d too small", last)
		}
	}
}

// [end test]

func TestGrowthIsGeometric(t *testing.T) {
	var steps []int
	for _, c := range CapSteps(1_000_000) {
		if c >= 512 { // small sizes depend on allocator size classes
			steps = append(steps, c)
		}
	}
	t.Logf("capacities from 512 up: %v", steps)
	for i := 1; i < len(steps); i++ {
		if float64(steps[i]) < 1.2*float64(steps[i-1]) {
			t.Fatalf("capacity grew from %d to %d only", steps[i-1], steps[i])
		}
	}
}

// [start shrink]

func heapMiB() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc >> 20
}

func TestMapsDoNotShrink(t *testing.T) {
	Kept = nil // release what other tests left behind
	base := heapMiB()

	m := make(map[int][64]byte)
	for i := range 200_000 {
		m[i] = [64]byte{}
	}
	full := heapMiB() - base

	for k := range m {
		delete(m, k) // deleting during range is allowed
	}
	empty := heapMiB() - base
	runtime.KeepAlive(m)

	for i := range 200_000 {
		m[i] = [64]byte{}
	}
	clear(m) // removes every entry in one call
	cleared := heapMiB() - base
	runtime.KeepAlive(m)

	m = make(map[int][64]byte) // drop the old map entirely
	fresh := heapMiB() - base
	runtime.KeepAlive(m)

	t.Logf("full: %d MiB, deleted every key: %d MiB, after clear: %d MiB, replaced the map: %d MiB",
		full, empty, cleared, fresh)
	if empty < full/2 || cleared < full/2 || fresh > full/4 {
		t.Errorf("unexpected: full=%d empty=%d cleared=%d fresh=%d", full, empty, cleared, fresh)
	}
}

// [end shrink]
