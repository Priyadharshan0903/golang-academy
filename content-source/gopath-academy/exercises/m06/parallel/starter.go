//go:build !solution

// Package parallel practices goroutines with sync.WaitGroup, sync.Mutex
// and sync/atomic. Run the tests with -race.
package parallel

// Sum adds up nums using the given number of goroutines. Each worker
// sums one contiguous part of the slice; the partial results are then
// combined. workers is at least 1 and may be larger than len(nums).
func Sum(nums []int, workers int) int {
	// TODO: implement me.
	return 0
}

// Tally is a concurrency-safe word counter.
// Its zero value must be ready to use.
type Tally struct {
	// TODO: add a mutex and a map.
}

// Add increments the count for word.
func (t *Tally) Add(word string) {
	// TODO: implement me.
}

// Count returns the count for word.
func (t *Tally) Count(word string) int {
	// TODO: implement me.
	return 0
}

// Snapshot returns a copy of all counts. Changing the returned map
// must not affect the Tally.
func (t *Tally) Snapshot() map[string]int {
	// TODO: implement me.
	return nil
}
