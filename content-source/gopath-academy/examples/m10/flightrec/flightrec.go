//go:build go1.25

// Package flightrec holds the flight recorder example for lesson 10.3.
package flightrec

import (
	"bytes"
	"runtime/trace"
	"time"
)

// [start recorder]

// SlowRequestRecorder keeps the last few seconds of execution trace in
// memory and saves a snapshot only when something interesting happens,
// such as a request that took far longer than usual.
type SlowRequestRecorder struct {
	fr        *trace.FlightRecorder
	threshold time.Duration
	Snapshots [][]byte
}

func NewSlowRequestRecorder(threshold time.Duration) (*SlowRequestRecorder, error) {
	fr := trace.NewFlightRecorder(trace.FlightRecorderConfig{
		MinAge:   5 * time.Second, // keep at least the last 5s of events
		MaxBytes: 4 << 20,         // but at most about 4 MiB
	})
	if err := fr.Start(); err != nil {
		return nil, err
	}
	return &SlowRequestRecorder{fr: fr, threshold: threshold}, nil
}

// Observe is called after each request with its duration.
func (r *SlowRequestRecorder) Observe(d time.Duration) error {
	if d < r.threshold || len(r.Snapshots) > 0 { // one snapshot is enough here
		return nil
	}
	var buf bytes.Buffer
	if _, err := r.fr.WriteTo(&buf); err != nil { // the trace leading up to now
		return err
	}
	r.Snapshots = append(r.Snapshots, buf.Bytes())
	return nil
}

func (r *SlowRequestRecorder) Stop() { r.fr.Stop() }

// [end recorder]
