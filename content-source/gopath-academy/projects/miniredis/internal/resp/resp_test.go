package resp

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
)

func TestReadCommand(t *testing.T) {
	in := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$9\r\nAna Maria\r\n" + // value contains a space
		"PING\r\n" + // inline form
		"*1\r\n$0\r\n\r\n" // empty argument
	r := NewReader(strings.NewReader(in))

	want := [][]string{{"SET", "name", "Ana Maria"}, {"PING"}, {""}}
	for i, w := range want {
		got, err := r.ReadCommand()
		if err != nil || !slices.Equal(got, w) {
			t.Fatalf("command %d = %q, %v; want %q", i, got, err, w)
		}
	}
	if _, err := r.ReadCommand(); !errors.Is(err, io.EOF) {
		t.Errorf("after the last command err = %v, want io.EOF", err)
	}
}

func TestBinarySafe(t *testing.T) {
	payload := "line1\r\nline2\x00end"
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.Array([]string{"SET", "k", payload})
	w.Flush()

	got, err := NewReader(&buf).ReadCommand()
	if err != nil || got[2] != payload {
		t.Errorf("round trip lost data: %q, %v", got, err)
	}
}

func TestMalformed(t *testing.T) {
	cases := map[string]string{
		"bad array length": "*x\r\n",
		"negative length":  "*-5\r\n",
		"too many args":    "*99999\r\n",
		"not a bulk":       "*1\r\n:5\r\n",
		"missing CRLF":     "*1\r\n$3\r\nGETXX",
	}
	for name, in := range cases {
		_, err := NewReader(strings.NewReader(in)).ReadCommand()
		if !errors.Is(err, ErrProtocol) {
			t.Errorf("%s: err = %v, want ErrProtocol", name, err)
		}
	}
}

func TestTruncated(t *testing.T) {
	_, err := NewReader(strings.NewReader("*2\r\n$3\r\nGET\r\n")).ReadCommand()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("err = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.Simple("OK")
	w.Error("ERR wrong type")
	w.Int(-7)
	w.Bulk("hi")
	w.Null()
	w.Flush()
	want := "+OK\r\n-ERR wrong type\r\n:-7\r\n$2\r\nhi\r\n$-1\r\n"
	if buf.String() != want {
		t.Errorf("got %q\nwant %q", buf.String(), want)
	}
}

func FuzzReadCommand(f *testing.F) {
	f.Add("*1\r\n$4\r\nPING\r\n")
	f.Add("PING\r\n")
	f.Fuzz(func(t *testing.T, in string) {
		r := NewReader(strings.NewReader(in))
		for range 10 {
			if _, err := r.ReadCommand(); err != nil {
				return // any error is fine; panics and hangs are not
			}
		}
	})
}
