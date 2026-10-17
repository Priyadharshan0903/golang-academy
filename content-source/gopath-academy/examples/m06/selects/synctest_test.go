//go:build go1.25

package selects

import (
	"testing"
	"testing/synctest"
	"time"
)

// [start synctest]

// Inside synctest.Test, time is virtual: sleeping 10 minutes takes
// no real time, and timeouts become deterministic (Go 1.25+).
func TestTimeoutWithVirtualTime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		select {
		case <-slowLookup(): // the goroutine sleeps 200ms of fake time
			t.Fatal("lookup should not win")
		case <-time.After(20 * time.Millisecond):
		}
		if elapsed := time.Since(start); elapsed != 20*time.Millisecond {
			t.Errorf("elapsed %v, want exactly 20ms of virtual time", elapsed)
		}
		time.Sleep(time.Second) // let the slow goroutine finish before the bubble ends
	})
}

// [end synctest]
