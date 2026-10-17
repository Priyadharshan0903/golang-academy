//go:build solution

// Package parallel practices goroutines with sync.WaitGroup, sync.Mutex
// and sync/atomic.
package parallel

import (
	"maps"
	"sync"
)

// Sum adds up nums using the given number of goroutines.
func Sum(nums []int, workers int) int {
	workers = max(1, min(workers, len(nums)))
	partial := make([]int, workers) // one slot per worker: no lock needed
	chunk := (len(nums) + workers - 1) / workers

	var wg sync.WaitGroup
	for w := range workers {
		start := w * chunk
		end := min(start+chunk, len(nums))
		if start >= end {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, n := range nums[start:end] {
				partial[w] += n
			}
		}()
	}
	wg.Wait()

	total := 0
	for _, p := range partial {
		total += p
	}
	return total
}

// Tally is a concurrency-safe word counter.
type Tally struct {
	mu     sync.Mutex
	counts map[string]int
}

// Add increments the count for word.
func (t *Tally) Add(word string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.counts == nil {
		t.counts = make(map[string]int)
	}
	t.counts[word]++
}

// Count returns the count for word.
func (t *Tally) Count(word string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.counts[word]
}

// Snapshot returns a copy of all counts.
func (t *Tally) Snapshot() map[string]int {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := maps.Clone(t.counts)
	if out == nil {
		out = map[string]int{}
	}
	return out
}
