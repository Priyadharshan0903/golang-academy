package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStdin(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"-n", "2"}, strings.NewReader("b a b c b a"), &out, &errOut)
	if code != 0 {
		t.Fatalf("exit %d, stderr: %s", code, errOut.String())
	}
	want := "      3  b\n      2  a\n"
	if out.String() != want {
		t.Errorf("output =\n%q\nwant\n%q", out.String(), want)
	}
}

func TestRunFiles(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "one.txt")
	f2 := filepath.Join(dir, "two.txt")
	os.WriteFile(f1, []byte("go go rust"), 0o644)
	os.WriteFile(f2, []byte("go zig"), 0o644)

	var out, errOut bytes.Buffer
	if code := run([]string{"-n", "1", f1, f2}, nil, &out, &errOut); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut.String())
	}
	if out.String() != "      3  go\n" {
		t.Errorf("output = %q", out.String())
	}
}

func TestRunMissingFile(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"/does/not/exist.txt"}, nil, &out, &errOut)
	if code != 1 || !strings.Contains(errOut.String(), "exist.txt") {
		t.Errorf("code=%d stderr=%q", code, errOut.String())
	}
}
