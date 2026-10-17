package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestReadURLs(t *testing.T) {
	got, err := readURLs(strings.NewReader("# a comment\n\n https://example.test/a \nhttps://example.test/b\n"))
	if err != nil || len(got) != 2 || got[0] != "https://example.test/a" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}
func TestReadURLsRejectsOversizedLine(t *testing.T) {
	if _, err := readURLs(strings.NewReader(strings.Repeat("a", 70*1024))); err == nil {
		t.Fatal("accepted an oversized input line")
	}
}
func TestReadURLsRejectsTooManyURLs(t *testing.T) {
	if _, err := readURLs(strings.NewReader(strings.Repeat("https://example.test\n", 1001))); err == nil {
		t.Fatal("accepted too many URLs")
	}
}
func TestCLIUsage(t *testing.T) {
	for _, args := range [][]string{nil, {"-workers", "0", "https://example.test"}, {"-total", "0s", "https://example.test"}, {"-unknown"}} {
		var out, diagnostics bytes.Buffer
		if code := run(context.Background(), args, strings.NewReader(""), &out, &diagnostics); code != 2 {
			t.Fatalf("args=%v code=%d", args, code)
		}
		if diagnostics.Len() == 0 {
			t.Fatal("invalid usage must have diagnostics")
		}
	}
}
func TestHelp(t *testing.T) {
	var out, diagnostics bytes.Buffer
	if code := run(context.Background(), []string{"-help"}, strings.NewReader(""), &out, &diagnostics); code != 0 {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(diagnostics.String(), "workers") {
		t.Fatal("help did not explain workers")
	}
}
func TestCanceledRunReportsFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, diagnostics bytes.Buffer
	if code := run(ctx, []string{"https://example.test"}, strings.NewReader(""), &out, &diagnostics); code != 1 {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(diagnostics.String(), "canceled") {
		t.Fatalf("diagnostics=%q", diagnostics.String())
	}
}
