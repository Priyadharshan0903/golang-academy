package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// jsonq runs the program in-process, as main would, and returns the
// exit code and both outputs.
func jsonq(stdin string, args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run(context.Background(), args, strings.NewReader(stdin), &out, &errOut)
	return code, out.String(), errOut.String()
}

const user = `{"name":"Ada","emails":["ada@example.com","countess@example.org"],"age":36}`

// [start example]

func Example() {
	code, out, _ := jsonq(user, "keys", "-sort")
	fmt.Print(out)
	fmt.Println("exit", code)

	_, out, _ = jsonq(user, "get", "-raw", "emails.1")
	fmt.Print(out)

	code, _, errOut := jsonq(user, "get", "emails.5")
	fmt.Print(errOut)
	fmt.Println("exit", code)

	code, _, errOut = jsonq(user, "pretty", "-indent", "9")
	fmt.Print(errOut)
	fmt.Println("exit", code)
	// Output:
	// age
	// emails
	// name
	// exit 0
	// countess@example.org
	// jsonq: stdin: no index "5" in an array of 2
	// exit 1
	// jsonq: usage error: -indent must be between 0 and 8
	// exit 2
}

// [end example]

func TestUsage(t *testing.T) {
	code, _, errOut := jsonq("", "frobnicate")
	if code != 2 || !strings.Contains(errOut, `unknown command "frobnicate"`) || !strings.Contains(errOut, "pretty") {
		t.Fatalf("code %d, stderr:\n%s", code, errOut)
	}
	code, _, errOut = jsonq("", "get", "-h")
	if code != 0 || !strings.Contains(errOut, "-raw") {
		t.Fatalf("-h: code %d, stderr:\n%s", code, errOut)
	}
	code, _, errOut = jsonq("", "keys", "-nope")
	if code != 2 || !strings.Contains(errOut, "flag provided but not defined: -nope") {
		t.Fatalf("bad flag: code %d, stderr:\n%s", code, errOut)
	}
}

func TestFilesAndStreams(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.json")
	os.WriteFile(a, []byte(`{"x":1}`+"\n"+`{"x":2}`), 0o644) // two documents in one file

	code, out, _ := jsonq("", "get", "x", a)
	if code != 0 || out != "1\n2\n" {
		t.Fatalf("code %d out %q", code, out)
	}
	code, _, errOut := jsonq("", "keys", filepath.Join(dir, "missing.json"))
	if code != 1 || !strings.Contains(errOut, "missing.json") {
		t.Fatalf("missing file: code %d, stderr %q", code, errOut)
	}
	code, _, errOut = jsonq(`{"x":`, "pretty")
	if code != 1 || !strings.Contains(errOut, "invalid JSON") {
		t.Fatalf("bad JSON: code %d, stderr %q", code, errOut)
	}
}

func TestCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, errOut bytes.Buffer
	if code := run(ctx, []string{"pretty"}, strings.NewReader(user), &out, &errOut); code != 1 || out.Len() != 0 {
		t.Fatalf("code %d out %q", code, out.String())
	}
}
