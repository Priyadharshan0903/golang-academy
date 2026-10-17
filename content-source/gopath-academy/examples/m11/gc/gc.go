// Package gc holds the examples for lesson 11.3.
package gc

import (
	"runtime"
	"runtime/metrics"
	"slices"
	"weak"
)

// [start cycles]

var sink []byte

// Result describes one run of Churn.
type Result struct {
	Cycles     uint64 // GC cycles that ran during the churn
	MedianGoal uint64 // the typical heap goal during the churn, in bytes
}

// Churn keeps live bytes reachable, then allocates churn bytes of
// short-lived garbage in 4 KiB pieces, and reports how often the GC
// ran. GOGC and GOMEMLIMIT come from the environment, so run it in a
// fresh process for each setting (see gcdemo).
func Churn(live, churn int) Result {
	keep := make([][]byte, live/4096)
	for i := range keep {
		keep[i] = make([]byte, 4096)
	}
	runtime.GC()

	start := read("/gc/cycles/total:gc-cycles")
	var goals []uint64
	for i := range churn / 4096 {
		sink = make([]byte, 4096)
		if i%256 == 0 {
			goals = append(goals, read("/gc/heap/goal:bytes"))
		}
	}
	cycles := read("/gc/cycles/total:gc-cycles") - start
	runtime.KeepAlive(keep)

	slices.Sort(goals)
	return Result{Cycles: cycles, MedianGoal: goals[len(goals)/2]}
}

func read(name string) uint64 {
	s := []metrics.Sample{{Name: name}}
	metrics.Read(s)
	return s[0].Value.Uint64()
}

// [end cycles]

// [start weak]

// Blob is big enough to get its own heap allocation.
type Blob struct{ data [1024]byte }

// newWeakBlob allocates a Blob and returns only a weak pointer to it.
//
//go:noinline
func newWeakBlob() weak.Pointer[Blob] {
	return weak.Make(&Blob{})
}

// alive reports whether w's target still exists. Value returns an
// ordinary (strong) pointer, which keeps the target alive for as long
// as it is in use, so keep that use short and separate.
//
//go:noinline
func alive(w weak.Pointer[Blob]) bool { return w.Value() != nil }

// WeakDemo shows that a weak pointer does not keep its target alive.
func WeakDemo() (aliveBefore, aliveAfter bool) {
	w := newWeakBlob()
	aliveBefore = alive(w) // nothing but w refers to the Blob
	runtime.GC()
	aliveAfter = alive(w)
	return aliveBefore, aliveAfter
}

// [end weak]

// [start cleanup]

// CleanupDemo attaches a cleanup function to an object. The cleanup
// runs, on a separate goroutine, some time after the object becomes
// unreachable.
func CleanupDemo(done chan<- string) {
	b := &Blob{}
	runtime.AddCleanup(b, func(name string) { done <- name }, "blob #1")
	runtime.KeepAlive(b)
}

// [end cleanup]
