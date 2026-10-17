//go:build !solution

// Package rle practices fuzzing with a run-length encoder. The tests
// include a fuzz test; run it with:
//
//	go test -fuzz=FuzzRoundTrip -fuzztime=20s ./m09/rle
package rle

import "errors"

// ErrCorrupt is returned by Decode for invalid input.
var ErrCorrupt = errors.New("rle: corrupt data")

// Encode compresses data as pairs of bytes: (count, value), where count
// is 1 to 255. A run longer than 255 is split into several pairs.
//
//	Encode([]byte("aaab")) == []byte{3, 'a', 1, 'b'}
//
// Encode(nil) and Encode([]byte{}) return an empty (or nil) slice.
func Encode(data []byte) []byte {
	// TODO: implement me.
	return nil
}

// Decode reverses Encode. It returns ErrCorrupt if the input has an odd
// length or contains a pair with count 0.
func Decode(enc []byte) ([]byte, error) {
	// TODO: implement me.
	return nil, nil
}
