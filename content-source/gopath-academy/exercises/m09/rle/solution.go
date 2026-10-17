//go:build solution

// Package rle practices fuzzing with a run-length encoder.
package rle

import "errors"

// ErrCorrupt is returned by Decode for invalid input.
var ErrCorrupt = errors.New("rle: corrupt data")

// Encode compresses data as (count, value) byte pairs.
func Encode(data []byte) []byte {
	var out []byte
	for i := 0; i < len(data); {
		v := data[i]
		n := 1
		for i+n < len(data) && data[i+n] == v && n < 255 {
			n++
		}
		out = append(out, byte(n), v)
		i += n
	}
	return out
}

// Decode reverses Encode.
func Decode(enc []byte) ([]byte, error) {
	if len(enc)%2 != 0 {
		return nil, ErrCorrupt
	}
	var out []byte
	for i := 0; i < len(enc); i += 2 {
		n, v := int(enc[i]), enc[i+1]
		if n == 0 {
			return nil, ErrCorrupt
		}
		for range n {
			out = append(out, v)
		}
	}
	return out, nil
}
