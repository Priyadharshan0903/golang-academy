package goroutines

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// [start start]

func Example_startGoroutines() {
	var wg sync.WaitGroup
	results := make([]string, 3) // each goroutine writes its own index

	for i := range 3 {
		wg.Add(1)
		go func() { // "go" starts the call in a new goroutine and returns at once
			defer wg.Done()
			results[i] = fmt.Sprintf("worker %d done", i)
		}()
	}

	wg.Wait() // block until every Done has been called
	for _, r := range results {
		fmt.Println(r)
	}

	// Output:
	// worker 0 done
	// worker 1 done
	// worker 2 done
}

// [end start]

// [start cheap]

func Example_cheap() {
	var started, done sync.WaitGroup
	release := make(chan struct{})
	for range 10_000 {
		started.Add(1)
		done.Add(1)
		go func() {
			defer done.Done()
			started.Done()
			<-release // park until released
		}()
	}
	started.Wait() // all 10,000 are running (or parked) now

	fmt.Println(runtime.NumGoroutine() >= 10_000)
	close(release) // wakes every goroutine at once
	done.Wait()

	// Output:
	// true
}

// [end cheap]

// [start mainexit]

// In a real program, main returning ends the process immediately,
// even if goroutines are still running. This helper shows the effect
// without ending the test binary: work that is not waited for is lost.
func runWithoutWaiting() []string {
	var mu sync.Mutex
	var log []string
	go func() {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		log = append(log, "late work")
		mu.Unlock()
	}()
	mu.Lock()
	defer mu.Unlock()
	return append([]string(nil), log...) // returns before the goroutine runs
}

func Example_notWaiting() {
	fmt.Println(len(runWithoutWaiting()))

	// Output:
	// 0
}

// [end mainexit]
