package coverage

import (
	"fmt"
	"os"
	"testing"
)

// [start testmain]

// TestMain runs once per test binary, around all tests in the package.
// Use it for expensive shared setup such as starting a database
// container. Most packages do not need it.
func TestMain(m *testing.M) {
	setup := func() { /* start shared resources */ }
	teardown := func() { /* stop them */ }

	setup()
	code := m.Run() // runs the tests
	teardown()
	if code != 0 {
		fmt.Fprintln(os.Stderr, "tests failed")
	}
	os.Exit(code)
}

// [end testmain]
