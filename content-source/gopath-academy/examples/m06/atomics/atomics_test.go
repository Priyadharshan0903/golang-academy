package atomics

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// [start counter]

type Stats struct {
	requests atomic.Int64 // zero value is 0 and ready to use
	healthy  atomic.Bool
}

func Example_atomicTypes() {
	var s Stats
	s.healthy.Store(true)

	var wg sync.WaitGroup
	for range 1000 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.requests.Add(1) // atomic read-modify-write, no lock needed
		}()
	}
	wg.Wait()
	fmt.Println(s.requests.Load(), s.healthy.Load())

	// Output:
	// 1000 true
}

// [end counter]

// [start cas]

// claim lets exactly one goroutine win, using compare-and-swap.
func Example_compareAndSwap() {
	var owner atomic.Int32 // 0 means "unclaimed"
	var wg sync.WaitGroup
	var winners atomic.Int32

	for id := int32(1); id <= 10; id++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if owner.CompareAndSwap(0, id) { // only succeeds if still 0
				winners.Add(1)
			}
		}()
	}
	wg.Wait()
	fmt.Println(winners.Load(), owner.Load() != 0)

	// Output:
	// 1 true
}

// [end cas]

// [start pointer]

type Settings struct {
	MaxConns int
	Region   string
}

// current holds an immutable snapshot that can be swapped atomically.
// Readers never see a half-updated Settings.
var current atomic.Pointer[Settings]

func Example_atomicPointer() {
	current.Store(&Settings{MaxConns: 10, Region: "ap-south-1"})

	cfg := current.Load() // a consistent snapshot
	fmt.Println(cfg.MaxConns, cfg.Region)

	// Publish new settings by storing a NEW value; never mutate the old one.
	current.Store(&Settings{MaxConns: 50, Region: "ap-south-1"})
	fmt.Println(current.Load().MaxConns, cfg.MaxConns)

	// Output:
	// 10 ap-south-1
	// 50 10
}

// [end pointer]
