package poll

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWaitForImmediate(t *testing.T) {
	calls := 0
	err := WaitFor(context.Background(), time.Hour, func() bool { calls++; return true })
	if err != nil || calls != 1 {
		t.Errorf("err=%v calls=%d; cond should be checked once, immediately", err, calls)
	}
}

func TestWaitForEventually(t *testing.T) {
	var n atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := WaitFor(ctx, time.Millisecond, func() bool { return n.Add(1) >= 5 })
	if err != nil {
		t.Fatalf("WaitFor: %v", err)
	}
	if n.Load() != 5 {
		t.Errorf("cond called %d times, want exactly 5", n.Load())
	}
}

func TestWaitForTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := WaitFor(ctx, 2*time.Millisecond, func() bool { return false })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
	if time.Since(start) > 500*time.Millisecond {
		t.Error("WaitFor did not stop promptly at the deadline")
	}
}

func TestRetrySucceeds(t *testing.T) {
	calls := 0
	err := Retry(context.Background(), 5, time.Millisecond, func(context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Errorf("err=%v calls=%d, want nil and 3", err, calls)
	}
}

func TestRetryGivesUp(t *testing.T) {
	boom := errors.New("boom")
	calls := 0
	err := Retry(context.Background(), 3, time.Millisecond, func(context.Context) error {
		calls++
		return boom
	})
	if !errors.Is(err, boom) || calls != 3 {
		t.Errorf("err=%v calls=%d, want boom after 3 calls", err, calls)
	}
}

func TestRetryBackoffDoubles(t *testing.T) {
	var times []time.Time
	Retry(context.Background(), 3, 20*time.Millisecond, func(context.Context) error {
		times = append(times, time.Now())
		return errors.New("fail")
	})
	if len(times) != 3 {
		t.Fatalf("got %d calls", len(times))
	}
	first, second := times[1].Sub(times[0]), times[2].Sub(times[1])
	if first < 20*time.Millisecond || second < 40*time.Millisecond {
		t.Errorf("waits were %v and %v, want at least 20ms then 40ms", first, second)
	}
}

func TestRetryRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	calls := 0
	err := Retry(ctx, 10, 100*time.Millisecond, func(context.Context) error {
		calls++
		return errors.New("fail")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
	if calls != 1 || time.Since(start) > 90*time.Millisecond {
		t.Errorf("calls=%d after %v; Retry should stop waiting as soon as ctx is done", calls, time.Since(start))
	}
}
