package racedetector

import (
	"sync"
	"sync/atomic"
	"testing"
)

// [start fixed]

// The three standard ways to fix a data race on a counter. Run with
// go test -race to confirm that none of them is reported.

func TestFixedWithMutex(t *testing.T) {
	var (
		mu sync.Mutex
		n  int
		wg sync.WaitGroup
	)
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			n++
			mu.Unlock()
		}()
	}
	wg.Wait()
	if n != 100 {
		t.Fatalf("n = %d", n)
	}
}

func TestFixedWithAtomic(t *testing.T) {
	var n atomic.Int64
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.Add(1)
		}()
	}
	wg.Wait()
	if n.Load() != 100 {
		t.Fatalf("n = %d", n.Load())
	}
}

func TestFixedWithChannel(t *testing.T) {
	increments := make(chan int)
	done := make(chan int)
	go func() { // one goroutine owns the counter
		n := 0
		for d := range increments {
			n += d
		}
		done <- n
	}()

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			increments <- 1
		}()
	}
	wg.Wait()
	close(increments)
	if n := <-done; n != 100 {
		t.Fatalf("n = %d", n)
	}
}

// [end fixed]
