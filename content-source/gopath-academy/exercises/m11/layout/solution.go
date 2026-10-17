//go:build solution

// Package layout practices struct field ordering.
package layout

// Event is kept in memory by the million, so every byte counts.
//
// Fields are ordered by alignment, largest first: the 8-byte-aligned
// fields (string, int64, float64), then int32, int16 and the 1-byte
// fields. The data adds up to 41 bytes, and the struct's size must be
// a multiple of its alignment (8), so 48 bytes is the minimum.
type Event struct {
	Name     string  // 16 bytes, offset 0
	ID       int64   // 8 bytes, offset 16
	Score    float64 // 8 bytes, offset 24
	Priority int32   // 4 bytes, offset 32
	Retries  int16   // 2 bytes, offset 36
	Active   bool    // 1 byte, offset 38
	Kind     uint8   // 1 byte, offset 39
	Deleted  bool    // 1 byte, offset 40, then 7 bytes of padding
}
