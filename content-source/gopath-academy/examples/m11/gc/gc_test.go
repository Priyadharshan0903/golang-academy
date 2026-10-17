package gc

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// [start test]

// TestHelperChurn is not a real test: TestGOGC runs the test binary
// again with GC_HELPER=1 and different GOGC settings, so every setting
// starts from a fresh runtime.
func TestHelperChurn(t *testing.T) {
	if os.Getenv("GC_HELPER") != "1" {
		t.Skip("helper process only")
	}
	json.NewEncoder(os.Stdout).Encode(Churn(16<<20, 512<<20))
	os.Exit(0)
}

func churnWith(t *testing.T, env ...string) Result {
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperChurn$")
	cmd.Env = append(os.Environ(), append(env, "GC_HELPER=1")...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("helper %v: %v", env, err)
	}
	var r Result
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("helper output %q: %v", out, err)
	}
	return r
}

func TestGOGC(t *testing.T) {
	cycles := map[string]uint64{}
	for _, gogc := range []string{"50", "100", "200", "400"} {
		r := churnWith(t, "GOGC="+gogc, "GOMEMLIMIT=off")
		t.Logf("GOGC=%-4s cycles=%-3d heap goal about %3d MiB", gogc, r.Cycles, r.MedianGoal>>20)
		cycles[gogc] = r.Cycles
	}
	// Exact counts vary with machine load, so check only the trend:
	// eight times the GOGC must mean clearly fewer cycles.
	if cycles["400"]*2 > cycles["50"] {
		t.Errorf("GOGC=400 ran %d cycles, GOGC=50 ran %d: expected far fewer", cycles["400"], cycles["50"])
	}

	r := churnWith(t, "GOGC=off", "GOMEMLIMIT=64MiB")
	t.Logf("GOGC=off GOMEMLIMIT=64MiB: cycles=%-3d heap goal about %3d MiB", r.Cycles, r.MedianGoal>>20)
	if r.Cycles == 0 || r.MedianGoal > 64<<20 {
		t.Errorf("memory limit not respected: %+v", r)
	}
}

// [end test]

func TestWeak(t *testing.T) {
	before, after := WeakDemo()
	if !before || after {
		t.Fatalf("aliveBefore=%v aliveAfter=%v, want true false", before, after)
	}
}

func TestCleanup(t *testing.T) {
	done := make(chan string, 1)
	CleanupDemo(done)
	runtime.GC()
	select {
	case name := <-done:
		t.Logf("cleanup ran for %s", name)
	case <-time.After(5 * time.Second):
		t.Fatal("cleanup did not run")
	}
}
