package pmap

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestOrderPreserved(t *testing.T) {
	inputs := []int{5, 1, 4, 2, 3}
	got, err := Map(context.Background(), inputs, 3, func(_ context.Context, n int) (string, error) {
		time.Sleep(time.Duration(n) * time.Millisecond) // finish out of order
		return strconv.Itoa(n * 10), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"50", "10", "40", "20", "30"}) {
		t.Errorf("got %v", got)
	}
}

func TestEmptyInput(t *testing.T) {
	got, err := Map(context.Background(), []int{}, 4, func(context.Context, int) (int, error) { return 0, nil })
	if err != nil || len(got) != 0 {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestConcurrencyLimit(t *testing.T) {
	var running, peak atomic.Int32
	_, err := Map(context.Background(), make([]int, 30), 4, func(context.Context, int) (int, error) {
		n := running.Add(1)
		for p := peak.Load(); n > p && !peak.CompareAndSwap(p, n); p = peak.Load() {
		}
		time.Sleep(2 * time.Millisecond)
		running.Add(-1)
		return 0, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if peak.Load() > 4 {
		t.Errorf("peak concurrency %d, want at most 4", peak.Load())
	}
	if peak.Load() < 2 {
		t.Errorf("peak concurrency %d, the work does not seem to run in parallel", peak.Load())
	}
}

func TestFirstErrorCancelsOthers(t *testing.T) {
	boom := errors.New("boom")
	var started, sawCancel atomic.Int32
	_, err := Map(context.Background(), make([]int, 100), 5, func(ctx context.Context, _ int) (int, error) {
		if started.Add(1) == 3 {
			return 0, boom
		}
		select {
		case <-ctx.Done():
			sawCancel.Add(1)
			return 0, ctx.Err()
		case <-time.After(time.Second):
			return 0, nil
		}
	})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want boom", err)
	}
	if sawCancel.Load() == 0 {
		t.Error("other calls were not cancelled")
	}
	if started.Load() == 100 {
		t.Error("Map kept starting work after the first error")
	}
}

func TestParentCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := Map(ctx, make([]int, 1000), 2, func(ctx context.Context, _ int) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(5 * time.Millisecond):
			return 0, nil
		}
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
}

func TestNoLeaks(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := range 20 {
		Map(context.Background(), make([]int, 50), 8, func(_ context.Context, n int) (int, error) {
			if i%2 == 0 {
				return 0, fmt.Errorf("fail")
			}
			return n, nil
		})
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Errorf("%d goroutines leaked", n-before)
	}
}
