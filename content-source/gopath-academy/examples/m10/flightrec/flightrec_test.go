//go:build go1.25

package flightrec

import (
	"testing"
	"time"
)

func TestSnapshotOnSlowRequest(t *testing.T) {
	r, err := NewSlowRequestRecorder(50 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Stop()

	for _, d := range []time.Duration{2 * time.Millisecond, 3 * time.Millisecond} {
		r.Observe(d)
	}
	if len(r.Snapshots) != 0 {
		t.Fatal("fast requests should not trigger a snapshot")
	}

	time.Sleep(20 * time.Millisecond) // some activity to record
	if err := r.Observe(200 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if len(r.Snapshots) != 1 || len(r.Snapshots[0]) == 0 {
		t.Fatalf("expected one non-empty trace snapshot, got %d", len(r.Snapshots))
	}
	// A real service would write the snapshot to a file and inspect it
	// with: go tool trace snapshot.trace
}
