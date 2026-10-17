package main

import (
	"path/filepath"
	"testing"

	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/aof"
	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/server"
	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/store"
)

// TestRestartKeepsData simulates a restart: write through a server with
// an AOF, then replay the AOF into a fresh store.
func TestRestartKeepsData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.aof")
	log, err := aof.Open(path, true)
	if err != nil {
		t.Fatal(err)
	}
	first := &server.Server{Store: store.New(), AOF: log}
	for _, cmd := range [][]string{{"SET", "a", "1"}, {"INCR", "a"}, {"SET", "b", "x"}, {"DEL", "b"}} {
		if err := first.Apply(cmd); err != nil {
			t.Fatal(err)
		}
	}
	log.Close()

	fresh := store.New()
	replayer := &server.Server{Store: fresh}
	if _, err := aof.Replay(path, replayer.Apply); err != nil {
		t.Fatal(err)
	}
	if v, ok := fresh.Get("a"); !ok || v != "2" {
		t.Errorf("a = %q, %v after restart; want 2", v, ok)
	}
	if fresh.Exists("b") != 0 {
		t.Error("b should stay deleted after restart")
	}
}
