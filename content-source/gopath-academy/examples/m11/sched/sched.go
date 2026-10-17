// Package sched holds the examples for lesson 11.1.
package sched

import (
	"crypto/sha256"
	"runtime"
	"runtime/metrics"
	"sync"
	"sync/atomic"
	"time"
)

// [start preempt]

// PreemptionDemo runs a goroutine with a tight loop on a single P
// (GOMAXPROCS=1) and checks that the calling goroutine still gets to
// run. It returns how long the round trip took.
func PreemptionDemo() time.Duration {
	prev := runtime.GOMAXPROCS(1) // one P: one goroutine runs Go code at a time
	defer runtime.GOMAXPROCS(prev)

	var stop atomic.Bool
	done := make(chan struct{})
	go func() {
		n := 0
		for !stop.Load() { // a tight loop: no blocking, no function calls
			n++
		}
		close(done)
	}()

	start := time.Now()
	time.Sleep(10 * time.Millisecond) // lets the spinner take the only P
	stop.Store(true)                  // we only get here because the spinner was preempted
	<-done
	return time.Since(start)
}

// [end preempt]

// [start metrics]

// SchedSnapshot reads a few scheduler metrics from runtime/metrics.
func SchedSnapshot() map[string]uint64 {
	samples := []metrics.Sample{
		{Name: "/sched/gomaxprocs:threads"},
		{Name: "/sched/goroutines:goroutines"},
	}
	metrics.Read(samples)
	out := map[string]uint64{}
	for _, s := range samples {
		if s.Value.Kind() == metrics.KindUint64 {
			out[s.Name] = s.Value.Uint64()
		}
	}
	return out
}

// [end metrics]

// [start parallel]

// HashAll hashes every block, splitting the blocks across workers
// goroutines. How many of them run at the same time is decided by
// GOMAXPROCS, not by the number of goroutines.
func HashAll(blocks [][]byte, workers int) [][32]byte {
	out := make([][32]byte, len(blocks))
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := w; i < len(blocks); i += workers {
				out[i] = sha256.Sum256(blocks[i])
			}
		}()
	}
	wg.Wait()
	return out
}

// [end parallel]
