package helpers

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// [start helper]

// writeFile is a test helper. t.Helper() makes failures point at the
// caller's line instead of this function.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name) // removed automatically after the test
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestLoadEnvFile(t *testing.T) {
	path := writeFile(t, ".env", "# comment\nDB_HOST = localhost\nDB_PORT=5432\n")
	got, err := LoadEnvFile(path)
	if err != nil {
		t.Fatal(err) // stop: nothing below makes sense without a result
	}
	if got["DB_HOST"] != "localhost" || got["DB_PORT"] != "5432" {
		t.Errorf("got %v", got)
	}
}

func TestLoadEnvFileError(t *testing.T) {
	path := writeFile(t, "bad.env", "OK=1\nbroken line\n")
	_, err := LoadEnvFile(path)
	if err == nil || err.Error() != "bad.env:2: missing '='" {
		t.Errorf("err = %v", err)
	}
}

// [end helper]

// [start golden]

var update = flag.Bool("update", false, "rewrite golden files with current output")

func TestRenderGolden(t *testing.T) {
	t.Setenv("APP_ENV", "test") // restored automatically after the test

	got := Render(map[string]string{"DB_PORT": "5432", "DB_HOST": "localhost", "LOG_LEVEL": "debug"})

	golden := filepath.Join("testdata", "render.golden")
	if *update {
		os.WriteFile(golden, []byte(got), 0o644)
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("output does not match %s (run with -update to accept)\n got:\n%s\nwant:\n%s", golden, got, want)
	}
}

// [end golden]

// [start cleanup]

func TestCleanupOrder(t *testing.T) {
	var log []string
	t.Run("inner", func(t *testing.T) {
		t.Cleanup(func() { log = append(log, "cleanup 1") })
		t.Cleanup(func() { log = append(log, "cleanup 2") }) // runs first: LIFO, like defer
		log = append(log, "test body")
	})
	want := "test body,cleanup 2,cleanup 1"
	if got := join(log); got != want {
		t.Errorf("order = %s, want %s", got, want)
	}
}

func join(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ","
		}
		out += v
	}
	return out
}

// [end cleanup]

// [start context]

func TestContext(t *testing.T) {
	ctx := t.Context() // Go 1.24+: cancelled just before Cleanup functions run
	select {
	case <-ctx.Done():
		t.Fatal("context should be live during the test")
	default:
	}
	t.Cleanup(func() {
		if ctx.Err() == nil {
			t.Error("context should be cancelled by the time cleanups run")
		}
	})
}

// [end context]
