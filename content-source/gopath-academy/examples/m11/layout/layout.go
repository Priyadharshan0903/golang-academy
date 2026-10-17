// Package layout holds the examples for lesson 11.4.
package layout

import "sync/atomic"

// [start structs]

// Wasteful orders its fields so that padding is needed after almost
// every one of them.
type Wasteful struct {
	Active bool  // 1 byte, then 7 bytes of padding so ID is 8-aligned
	ID     int64 // 8 bytes
	Kind   bool  // 1 byte, then 3 bytes of padding so Count is 4-aligned
	Count  int32 // 4 bytes
	Ready  bool  // 1 byte, then 7 bytes of padding to round up to 8
}

// Compact holds the same fields, largest alignment first.
type Compact struct {
	ID     int64
	Count  int32
	Active bool
	Kind   bool
	Ready  bool // 1 byte of padding follows to round 15 up to 16
}

// [end structs]

// [start sharing]

// Counters puts two hot counters next to each other: they share one
// 64-byte cache line.
type Counters struct {
	A atomic.Int64
	B atomic.Int64
}

// PaddedCounters keeps the counters 128 bytes apart. Some Intel CPUs
// fetch cache lines in pairs, so 64 bytes is not always enough.
type PaddedCounters struct {
	A atomic.Int64
	_ [120]byte
	B atomic.Int64
}

// [end sharing]
