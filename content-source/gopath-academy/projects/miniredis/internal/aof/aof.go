// Package aof implements an append-only file: every write command is
// appended to a log, and on startup the log is replayed to rebuild the
// data. It is the simplest durable persistence there is, and the same
// idea Redis' AOF mode uses.
package aof

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/miniredis/internal/resp"
)

// Log appends commands to a file. It is safe for concurrent use.
type Log struct {
	mu   sync.Mutex
	f    *os.File
	w    *resp.Writer
	sync bool // fsync after every append (slow but safest)
}

// Open opens or creates the log at path for appending.
func Open(path string, syncEveryWrite bool) (*Log, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open aof: %w", err)
	}
	return &Log{f: f, w: resp.NewWriter(f), sync: syncEveryWrite}, nil
}

// Append writes one command to the log.
func (l *Log) Append(args []string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.w.Array(args)
	if err := l.w.Flush(); err != nil {
		return err
	}
	if l.sync {
		return l.f.Sync()
	}
	return nil
}

// Sync flushes the file to disk. Call it periodically when not syncing
// every write (Redis' default policy is once per second).
func (l *Log) Sync() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.f.Sync()
}

// SyncEvery calls Sync on an interval until stop is closed.
func (l *Log) SyncEvery(interval time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			l.Sync()
		}
	}
}

// Close syncs and closes the file.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return errors.Join(l.f.Sync(), l.f.Close())
}

// Replay reads every command in the file at path and passes it to apply.
// A missing file is not an error. A truncated last command (the program
// crashed in the middle of a write) is ignored, like Redis does by default.
func Replay(path string, apply func(args []string) error) (int, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := resp.NewReader(bufio.NewReader(f))
	n := 0
	for {
		args, err := r.ReadCommand()
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return n, nil
		}
		if err != nil {
			return n, fmt.Errorf("replay aof command %d: %w", n+1, err)
		}
		if err := apply(args); err != nil {
			return n, fmt.Errorf("replay aof command %d: %w", n+1, err)
		}
		n++
	}
}
