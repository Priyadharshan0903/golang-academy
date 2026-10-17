//go:build go1.25

package synctesting

import (
	"testing"
	"testing/synctest"
	"time"
)

// [start test]

func TestTTLCacheExpiry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := NewTTLCache(time.Minute, 10*time.Second)
		defer c.Close()

		c.Set("session", "abc")

		time.Sleep(59 * time.Second) // virtual time: returns instantly
		if _, ok := c.Get("session"); !ok {
			t.Fatal("entry expired too early")
		}

		time.Sleep(2 * time.Second)
		if _, ok := c.Get("session"); ok {
			t.Fatal("entry should have expired")
		}

		// The sweeper ticks every 10s. Advance past the next tick, then
		// wait until every goroutine in the bubble is blocked again, so
		// the sweep has definitely finished before we look.
		time.Sleep(10 * time.Second)
		synctest.Wait()
		if n := c.Len(); n != 0 {
			t.Fatalf("sweeper left %d entries", n)
		}
	})
}

// [end test]
