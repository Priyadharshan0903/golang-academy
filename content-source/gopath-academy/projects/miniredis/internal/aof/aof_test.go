package aof

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestAppendAndReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")
	l, err := Open(path, true)
	if err != nil {
		t.Fatal(err)
	}
	cmds := [][]string{{"SET", "a", "1"}, {"SET", "b", "two words"}, {"DEL", "a"}}
	for _, c := range cmds {
		if err := l.Append(c); err != nil {
			t.Fatal(err)
		}
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}

	var got [][]string
	n, err := Replay(path, func(args []string) error { got = append(got, args); return nil })
	if err != nil || n != 3 {
		t.Fatalf("Replay = %d, %v", n, err)
	}
	for i := range cmds {
		if !slices.Equal(got[i], cmds[i]) {
			t.Errorf("command %d = %q, want %q", i, got[i], cmds[i])
		}
	}
}

func TestReplayMissingFile(t *testing.T) {
	n, err := Replay(filepath.Join(t.TempDir(), "none.aof"), func([]string) error { return nil })
	if n != 0 || err != nil {
		t.Errorf("Replay of a missing file = %d, %v", n, err)
	}
}

func TestReplayTruncatedTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "crash.aof")
	data := "*3\r\n$3\r\nSET\r\n$1\r\na\r\n$1\r\n1\r\n" + "*3\r\n$3\r\nSET\r\n$1\r\nb" // cut mid-write
	os.WriteFile(path, []byte(data), 0o644)

	n, err := Replay(path, func([]string) error { return nil })
	if err != nil || n != 1 {
		t.Errorf("Replay = %d, %v; want the 1 complete command and no error", n, err)
	}
}
