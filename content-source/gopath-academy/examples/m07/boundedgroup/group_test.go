package boundedgroup

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// [start use]

func Example_firstErrorCancelsOthers() {
	g, ctx := WithContext(context.Background())
	var cancelled atomic.Int32

	for i := range 5 {
		g.Go(func() error {
			if i == 2 {
				return fmt.Errorf("task %d failed", i)
			}
			select {
			case <-ctx.Done(): // the failure cancels the rest
				cancelled.Add(1)
				return ctx.Err()
			case <-time.After(time.Second):
				return nil
			}
		})
	}

	fmt.Println(g.Wait())
	fmt.Println(cancelled.Load())

	// Output:
	// task 2 failed
	// 4
}

// [end use]

// [start limit]

func Example_limit() {
	g, _ := WithContext(context.Background())
	g.SetLimit(3)

	var running, peak atomic.Int32
	for range 12 {
		g.Go(func() error {
			now := running.Add(1)
			for {
				p := peak.Load()
				if now <= p || peak.CompareAndSwap(p, now) {
					break
				}
			}
			time.Sleep(2 * time.Millisecond)
			running.Add(-1)
			return nil
		})
	}

	fmt.Println(g.Wait(), peak.Load() <= 3)

	// Output:
	// <nil> true
}

// [end limit]
