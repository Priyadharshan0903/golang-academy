// Package ratedemo shows golang.org/x/time/rate (lesson 7.5).
package ratedemo

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

// [start rate]

func Example_tokenBucket() {
	// 100 events per second on average, bursts of up to 5.
	lim := rate.NewLimiter(rate.Limit(100), 5)

	allowed := 0
	for range 10 {
		if lim.Allow() { // non-blocking: is a token available right now?
			allowed++
		}
	}
	fmt.Println("allowed immediately:", allowed)

	// Wait blocks until a token is available or ctx is done.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	start := time.Now()
	for range 3 {
		if err := lim.Wait(ctx); err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Println("waited at least 20ms:", time.Since(start) >= 20*time.Millisecond)

	// Output:
	// allowed immediately: 5
	// waited at least 20ms: true
}

// [end rate]
