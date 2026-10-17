package clicobra

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// [start test]

func execute(stdin string, args ...string) (string, error) {
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func Example() {
	out, _ := execute(`{"b":{"x":[1,2]},"a":true}`, "keys", "-s")
	fmt.Print(out)
	out, _ = execute(`{"b":{"x":[1,2]},"a":true}`, "get", "b", "--compact")
	fmt.Print(out)
	_, err := execute(`{}`, "get")
	fmt.Println(err)
	_, err = execute(`{}`, "lookup")
	fmt.Println(err)
	// Output:
	// a
	// b
	// {"x":[1,2]}
	// accepts 1 arg(s), received 0
	// unknown command "lookup" for "jsonq"
}

// [end test]

func TestHelpAndCompletion(t *testing.T) {
	out, err := execute("", "get", "--help")
	if err != nil || !strings.Contains(out, "jsonq get KEY") || !strings.Contains(out, "--compact") {
		t.Fatalf("help: %v\n%s", err, out)
	}
	// Cobra generates shell completion scripts for free.
	out, err = execute("", "completion", "bash")
	if err != nil || !strings.Contains(out, "bash completion") {
		t.Fatalf("completion: %v\n%.200s", err, out)
	}
	root, _ := execute("", "--help")
	t.Logf("jsonq --help:\n%s", root)
}
