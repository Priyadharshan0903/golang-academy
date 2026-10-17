package grep

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"
	"testing/iotest"
)

const logs = `INFO starting
ERROR disk full
INFO retrying
ERROR disk still full
DEBUG done`

func TestGrep(t *testing.T) {
	var out bytes.Buffer
	n, err := Grep(strings.NewReader(logs), &out, regexp.MustCompile(`^ERROR`), Options{})
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if want := "ERROR disk full\nERROR disk still full\n"; out.String() != want {
		t.Errorf("output %q, want %q", out.String(), want)
	}
}

func TestGrepOptions(t *testing.T) {
	var out bytes.Buffer
	n, _ := Grep(strings.NewReader(logs), &out, regexp.MustCompile(`INFO`), Options{LineNumbers: true, Invert: true})
	if want := "2:ERROR disk full\n4:ERROR disk still full\n5:DEBUG done\n"; out.String() != want || n != 3 {
		t.Errorf("n=%d output %q, want %q", n, out.String(), want)
	}
}

func TestGrepReadError(t *testing.T) {
	boom := errors.New("disk on fire")
	r := io.MultiReader(strings.NewReader("ERROR one\n"), iotest.ErrReader(boom))
	_, err := Grep(r, io.Discard, regexp.MustCompile(`ERROR`), Options{})
	if !errors.Is(err, boom) {
		t.Errorf("err = %v, want the reader's error", err)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("pipe closed") }

func TestGrepWriteError(t *testing.T) {
	_, err := Grep(strings.NewReader(logs), failWriter{}, regexp.MustCompile(`.`), Options{})
	if err == nil {
		t.Error("expected the writer's error to be returned")
	}
}

func TestCountingReader(t *testing.T) {
	cr := &CountingReader{R: iotest.HalfReader(strings.NewReader("hello, gopher"))}
	got, err := io.ReadAll(cr)
	if err != nil || string(got) != "hello, gopher" {
		t.Fatalf("ReadAll = %q, %v", got, err)
	}
	if cr.N != 13 {
		t.Errorf("N = %d, want 13", cr.N)
	}
	if err := iotest.TestReader(&CountingReader{R: strings.NewReader("abc")}, []byte("abc")); err != nil {
		t.Errorf("CountingReader breaks the io.Reader contract: %v", err)
	}
}

func TestCountingWithGrep(t *testing.T) {
	cr := &CountingReader{R: strings.NewReader(logs)}
	Grep(cr, io.Discard, regexp.MustCompile(`x`), Options{})
	if cr.N != int64(len(logs)) {
		t.Errorf("counted %d bytes, want %d", cr.N, len(logs))
	}
}
