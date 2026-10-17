//go:build go1.25

package ratelimit

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

// With synctest, Wait's real timers run on a virtual clock, so this
// test checks exact timing and still finishes instantly.
func TestWaitScheduleVirtualTime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		b := New(10, 1) // one token per 100ms
		start := time.Now()
		for range 5 {
			if err := b.Wait(context.Background()); err != nil {
				t.Fatal(err)
			}
		}
		// First event uses the burst token; the next 4 wait 100ms each.
		if got := time.Since(start); got != 400*time.Millisecond {
			t.Errorf("5 events took %v of virtual time, want exactly 400ms", got)
		}
	})
}
