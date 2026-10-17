//go:build go1.27

package leaks

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

// [start profile]

// The goroutineleak profile (Go 1.27+, experimental in 1.26) reports
// goroutines blocked on something no other goroutine can ever reach,
// found by the garbage collector. It is also served by net/http/pprof
// at /debug/pprof/goroutineleak.
func TestGoroutineLeakProfile(t *testing.T) {
	firstResultLeaky(3)
	time.Sleep(20 * time.Millisecond) // let the slow replicas block
	runtime.GC()

	var buf bytes.Buffer
	if err := pprof.Lookup("goroutineleak").WriteTo(&buf, 1); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "firstResultLeaky") {
		t.Errorf("expected the leaked goroutines in the profile, got:\n%s", buf.String())
	}
}

// [end profile]
