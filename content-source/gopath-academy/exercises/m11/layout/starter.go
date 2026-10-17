//go:build !solution

// Package layout practices struct field ordering.
package layout

// Event is kept in memory by the million, so every byte counts.
//
// TODO: reorder the fields so that unsafe.Sizeof(Event{}) is as small
// as possible. Do not add, remove, rename or change the type of any
// field. Hint: work out each field's size and alignment first.
type Event struct {
	Active   bool
	ID       int64
	Kind     uint8
	Score    float64
	Retries  int16
	Name     string
	Priority int32
	Deleted  bool
}
