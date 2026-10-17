package rle

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncode(t *testing.T) {
	if got := Encode([]byte("aaab")); !bytes.Equal(got, []byte{3, 'a', 1, 'b'}) {
		t.Errorf("Encode(aaab) = %v", got)
	}
	if got := Encode(nil); len(got) != 0 {
		t.Errorf("Encode(nil) = %v", got)
	}
	long := bytes.Repeat([]byte{'x'}, 300)
	if got := Encode(long); !bytes.Equal(got, []byte{255, 'x', 45, 'x'}) {
		t.Errorf("a 300-byte run should split into 255 + 45, got %v", got)
	}
}

func TestDecodeCorrupt(t *testing.T) {
	for _, bad := range [][]byte{{1}, {0, 'a'}, {2, 'a', 0, 'b'}} {
		if _, err := Decode(bad); !errors.Is(err, ErrCorrupt) {
			t.Errorf("Decode(%v) err = %v, want ErrCorrupt", bad, err)
		}
	}
}

// FuzzRoundTrip: for every input, Decode(Encode(x)) == x. Plain
// `go test` runs only the seeds; `go test -fuzz=FuzzRoundTrip` explores.
func FuzzRoundTrip(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("aaab"))
	f.Add(bytes.Repeat([]byte{0}, 600))
	f.Fuzz(func(t *testing.T, data []byte) {
		dec, err := Decode(Encode(data))
		if err != nil {
			t.Fatalf("Decode(Encode(%q)) error: %v", data, err)
		}
		if !bytes.Equal(dec, data) {
			t.Fatalf("round trip changed %q into %q", data, dec)
		}
	})
}

// FuzzDecode: arbitrary input must never panic, and anything Decode
// accepts must re-encode to something that decodes the same way.
func FuzzDecode(f *testing.F) {
	f.Add([]byte{2, 'a'})
	f.Fuzz(func(t *testing.T, enc []byte) {
		dec, err := Decode(enc)
		if err != nil {
			return
		}
		again, err := Decode(Encode(dec))
		if err != nil || !bytes.Equal(again, dec) {
			t.Fatalf("re-encoding %v broke the data", enc)
		}
	})
}
