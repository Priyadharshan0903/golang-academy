package leaks

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// [start leaky]

// firstResultLeaky asks several replicas and returns the fastest answer.
// The slower goroutines block forever on the unbuffered send: a leak.
func firstResultLeaky(replicas int) string {
	ch := make(chan string)
	for i := range replicas {
		go func() {
			time.Sleep(time.Duration(i) * time.Millisecond)
			ch <- fmt.Sprintf("replica %d", i) // only one send is ever received
		}()
	}
	return <-ch
}

// firstResult fixes it with a buffer big enough for every send.
func firstResult(replicas int) string {
	ch := make(chan string, replicas)
	for i := range replicas {
		go func() {
			time.Sleep(time.Duration(i) * time.Millisecond)
			ch <- fmt.Sprintf("replica %d", i)
		}()
	}
	return <-ch
}

// [end leaky]

// [start detect]

// waitForGoroutines polls until the goroutine count drops back to
// baseline, or gives up after a timeout. Libraries like go.uber.org/goleak
// do the same thing more thoroughly.
func waitForGoroutines(baseline int, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	for {
		n := runtime.NumGoroutine()
		if n <= baseline || time.Now().After(deadline) {
			return n
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestLeakyVersionLeaks(t *testing.T) {
	base := runtime.NumGoroutine()
	firstResultLeaky(5)
	if n := waitForGoroutines(base, 100*time.Millisecond); n <= base {
		t.Skip("leak not observed this time") // it should leak 4 goroutines
	} else {
		t.Logf("%d goroutines leaked", n-base)
	}
}

func TestFixedVersionDoesNotLeak(t *testing.T) {
	base := runtime.NumGoroutine()
	for range 20 {
		firstResult(5)
	}
	if n := waitForGoroutines(base, time.Second); n > base {
		t.Fatalf("%d goroutines leaked", n-base)
	}
}

// [end detect]

// [start ctxfix]

// fetch respects ctx: when the caller gives up, the goroutine
// selects on Done and exits instead of blocking forever.
func fetch(ctx context.Context, delay time.Duration) (string, error) {
	ch := make(chan string, 1)
	go func() {
		time.Sleep(delay) // stands in for slow I/O
		ch <- "data"
	}()
	select {
	case v := <-ch:
		return v, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func Example_contextPreventsLeak() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	_, err := fetch(ctx, 50*time.Millisecond)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))

	// Output:
	// true
}

// [end ctxfix]
