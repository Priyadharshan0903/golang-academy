package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{"default", nil, 0, "Hello, Gopher!\n", ""},
		{"flags", []string{"-name", "Ana", "-times", "2"}, 0, "Hello, Ana!\nHello, Ana!\n", ""},
		{"version", []string{"-version"}, 0, "greet dev (commit none, built unknown)\n", ""},
		{"bad times", []string{"-times", "0"}, 2, "", "-times must be at least 1"},
		{"unknown flag", []string{"-nope"}, 2, "", "flag provided but not defined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if got := stdout.String(); got != tt.wantOut {
				t.Errorf("stdout = %q, want %q", got, tt.wantOut)
			}
			if !strings.Contains(stderr.String(), tt.wantErr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tt.wantErr)
			}
		})
	}
}
