package ratelimiting

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// [start semaphore]

// A buffered channel is a counting semaphore: at most cap(sem)
// goroutines can hold a slot at the same time.
func Example_semaphore() {
	sem := make(chan struct{}, 2)
	var wg sync.WaitGroup
	var inFlight, peak atomic.Int32

	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			n := inFlight.Add(1)
			for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
			}
			time.Sleep(time.Millisecond)
			inFlight.Add(-1)
		}()
	}
	wg.Wait()
	fmt.Println(peak.Load() <= 2)

	// Output:
	// true
}

// [end semaphore]

// [start ticker]

// A ticker spaces out work: at most one request per interval.
func Example_tickerRateLimit() {
	requests := []string{"a", "b", "c", "d"}
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	for _, r := range requests {
		<-ticker.C // wait for the next tick before each request
		_ = r
	}
	elapsed := time.Since(start)
	fmt.Println(elapsed >= 15*time.Millisecond) // at least 3 full intervals between 4 requests

	// Output:
	// true
}

// [end ticker]

// [start backpressure]

// A bounded queue pushes back on producers: when the queue is full,
// Submit fails fast instead of letting memory grow without limit.
type Queue struct{ jobs chan string }

func NewQueue(size int) *Queue { return &Queue{jobs: make(chan string, size)} }

func (q *Queue) Submit(ctx context.Context, job string) error {
	select {
	case q.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("queue full, rejecting %q", job)
	}
}

func Example_backpressure() {
	q := NewQueue(2)
	ctx := context.Background()
	for _, j := range []string{"j1", "j2", "j3"} {
		if err := q.Submit(ctx, j); err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println(len(q.jobs))

	// Output:
	// queue full, rejecting "j3"
	// 2
}

// [end backpressure]
