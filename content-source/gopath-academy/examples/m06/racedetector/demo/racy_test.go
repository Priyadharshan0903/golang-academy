//go:build racedemo

// This file is excluded from normal builds. It contains a deliberate
// data race so you can see the race detector report it:
//
//	go test -race -tags racedemo ./examples/m06/racedetector/demo
package demo

import (
	"sync"
	"testing"
)

// [start racy]

func TestRacyCounter(t *testing.T) {
	n := 0
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n++ // DATA RACE: two goroutines read and write n unsynchronized
		}()
	}
	wg.Wait()
	t.Log("n =", n)
}

// [end racy]
