package todo

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type result struct {
	code           int
	stdout, stderr string
}

func todo(t *testing.T, env map[string]string, args ...string) result {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Run(args, func(k string) string { return env[k] }, &out, &errOut)
	return result{code, out.String(), errOut.String()}
}

func TestAddListDone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.json")
	f := []string{"-file", path}

	if r := todo(t, nil, append(f, "add", "buy", "milk")...); r.code != 0 {
		t.Fatalf("add: %+v", r)
	}
	todo(t, nil, append(f, "add", "write tests")...)
	todo(t, nil, append(f, "add", "ship it")...)

	if r := todo(t, nil, append(f, "done", "2")...); r.code != 0 {
		t.Fatalf("done 2: %+v", r)
	}
	r := todo(t, nil, append(f, "list")...)
	if r.code != 0 || r.stdout != "1. buy milk\n3. ship it\n" {
		t.Fatalf("list: code %d, stdout:\n%s", r.code, r.stdout)
	}
	r = todo(t, nil, append(f, "list", "-all")...)
	if r.stdout != "1. buy milk\n2. [x] write tests\n3. ship it\n" {
		t.Fatalf("list -all, stdout:\n%s", r.stdout)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the list was not saved to -file: %v", err)
	}
}

func TestFileFromEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "env.json")
	env := map[string]string{"TODO_FILE": path}
	todo(t, env, "add", "from env")
	if r := todo(t, env, "list"); r.stdout != "1. from env\n" {
		t.Fatalf("list via $TODO_FILE: %+v", r)
	}
	other := filepath.Join(t.TempDir(), "flag.json")
	if r := todo(t, env, "-file", other, "list"); r.stdout != "" {
		t.Fatalf("-file should win over $TODO_FILE: %+v", r)
	}
}

func TestExitCodes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.json")
	f := []string{"-file", path}
	todo(t, nil, append(f, "add", "one")...)

	tests := []struct {
		name string
		args []string
		code int
	}{
		{"no command", f, 2},
		{"unknown command", append(f, "remove", "1"), 2},
		{"bad global flag", []string{"-nope", "list"}, 2},
		{"bad list flag", append(f, "list", "-nope"), 2},
		{"add without text", append(f, "add"), 2},
		{"done without number", append(f, "done"), 2},
		{"done not a number", append(f, "done", "first"), 2},
		{"done out of range", append(f, "done", "7"), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := todo(t, nil, tt.args...)
			if r.code != tt.code {
				t.Errorf("exit code %d, want %d (stderr %q)", r.code, tt.code, r.stderr)
			}
			if !strings.HasPrefix(r.stderr, "todo: ") && !strings.Contains(r.stderr, "flag provided but not defined") {
				t.Errorf("stderr should explain the problem, got %q", r.stderr)
			}
		})
	}
}

func TestCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list.json")
	os.WriteFile(path, []byte("not json"), 0o644)
	if r := todo(t, nil, "-file", path, "list"); r.code != 1 {
		t.Fatalf("corrupt file: %+v, want exit code 1", r)
	}
}
