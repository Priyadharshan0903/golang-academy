package jobqueue

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/jobqueue/internal/pgtest"
)

func newClient(t *testing.T) *Client {
	t.Helper()
	c := New(pgtest.Pool(t))
	if err := c.Migrate(t.Context()); err != nil {
		t.Fatal(err)
	}
	return c
}

// start runs a worker in the background and returns a stop function
// that cancels it and waits for Run to return.
func start(t *testing.T, w *Worker) (stop func()) {
	t.Helper()
	if w.Logger == nil {
		w.Logger = slog.New(slog.DiscardHandler)
	}
	if w.PollInterval == 0 {
		w.PollInterval = 20 * time.Millisecond
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()
	return func() { cancel(); <-done }
}

// waitFor polls Counts until cond holds or 10 seconds pass.
func waitFor(t *testing.T, c *Client, queue string, cond func(map[string]int) bool) map[string]int {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		counts, err := c.Counts(t.Context(), queue)
		if err != nil {
			t.Fatal(err)
		}
		if cond(counts) {
			return counts
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out; counts = %v", counts)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// [start exactlyonce]

func TestEveryJobProcessedOnce(t *testing.T) {
	c := newClient(t)
	const n = 200
	for i := range n {
		if _, err := c.Enqueue(t.Context(), "emails", map[string]int{"n": i}, EnqueueOptions{}); err != nil {
			t.Fatal(err)
		}
	}

	var mu sync.Mutex
	seen := map[int64]int{}
	stop := start(t, &Worker{Client: c, Queue: "emails", Concurrency: 8, Handler: func(ctx context.Context, j Job) error {
		mu.Lock()
		seen[j.ID]++
		mu.Unlock()
		return nil
	}})
	begin := time.Now()
	waitFor(t, c, "emails", func(m map[string]int) bool { return m["done"] == n })
	stop()

	t.Logf("8 workers processed %d jobs in %v", n, time.Since(begin).Round(time.Millisecond))
	if len(seen) != n {
		t.Fatalf("processed %d distinct jobs, want %d", len(seen), n)
	}
	for id, times := range seen {
		if times != 1 {
			t.Errorf("job %d processed %d times", id, times)
		}
	}
}

// [end exactlyonce]

func TestRetryThenSucceed(t *testing.T) {
	c := newClient(t)
	id, _ := c.Enqueue(t.Context(), "q", "flaky", EnqueueOptions{})
	var attempts atomic.Int64
	stop := start(t, &Worker{Client: c, Queue: "q", Backoff: func(int) time.Duration { return 0 },
		Handler: func(ctx context.Context, j Job) error {
			if attempts.Add(1) < 3 {
				return errors.New("temporarily unavailable")
			}
			return nil
		}})
	defer stop()
	waitFor(t, c, "q", func(m map[string]int) bool { return m["done"] == 1 })

	var got int
	var lastErr string
	c.pool.QueryRow(t.Context(), `SELECT attempts, last_error FROM jobs WHERE id = $1`, id).Scan(&got, &lastErr)
	if got != 3 || lastErr != "temporarily unavailable" {
		t.Fatalf("attempts=%d last_error=%q, want 3 and the last failure", got, lastErr)
	}
}

func TestGivesUpAfterMaxAttempts(t *testing.T) {
	c := newClient(t)
	c.Enqueue(t.Context(), "q", "doomed", EnqueueOptions{MaxAttempts: 3})
	var calls atomic.Int64
	stop := start(t, &Worker{Client: c, Queue: "q", Backoff: func(int) time.Duration { return 0 },
		Handler: func(ctx context.Context, j Job) error {
			calls.Add(1)
			panic("bad payload") // panics are failures too
		}})
	defer stop()
	waitFor(t, c, "q", func(m map[string]int) bool { return m["failed"] == 1 })
	if calls.Load() != 3 {
		t.Fatalf("handler called %d times, want 3", calls.Load())
	}
}

// [start reclaim]

func TestReclaimAfterCrash(t *testing.T) {
	c := newClient(t)
	id, _ := c.Enqueue(t.Context(), "q", "x", EnqueueOptions{})

	// A worker claims the job and then "crashes": it never reports back.
	if _, err := c.claim(t.Context(), "q", 200*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	var gotAttempt atomic.Int64
	stop := start(t, &Worker{Client: c, Queue: "q", Handler: func(ctx context.Context, j Job) error {
		gotAttempt.Store(int64(j.Attempts))
		return nil
	}})
	defer stop()
	begin := time.Now()
	waitFor(t, c, "q", func(m map[string]int) bool { return m["done"] == 1 })
	t.Logf("job %d reclaimed after %v, on attempt %d", id, time.Since(begin).Round(10*time.Millisecond), gotAttempt.Load())
	if gotAttempt.Load() != 2 {
		t.Fatalf("attempt = %d, want 2", gotAttempt.Load())
	}
}

// [end reclaim]

// [start txenqueue]

func TestEnqueueInTransaction(t *testing.T) {
	c := newClient(t)
	ctx := t.Context()

	tx, _ := c.pool.Begin(ctx)
	EnqueueWith(ctx, tx, "q", "rolled back", EnqueueOptions{})
	tx.Rollback(ctx)

	tx, _ = c.pool.Begin(ctx)
	EnqueueWith(ctx, tx, "q", "committed", EnqueueOptions{})
	tx.Commit(ctx)

	counts, _ := c.Counts(ctx, "q")
	if counts["queued"] != 1 {
		t.Fatalf("counts = %v, want exactly the committed job", counts)
	}
}

// [end txenqueue]

func TestDelay(t *testing.T) {
	c := newClient(t)
	c.Enqueue(t.Context(), "q", "later", EnqueueOptions{Delay: 300 * time.Millisecond})
	var ranAt atomic.Int64
	enqueued := time.Now()
	stop := start(t, &Worker{Client: c, Queue: "q", Handler: func(ctx context.Context, j Job) error {
		ranAt.Store(int64(time.Since(enqueued)))
		return nil
	}})
	defer stop()
	waitFor(t, c, "q", func(m map[string]int) bool { return m["done"] == 1 })
	if d := time.Duration(ranAt.Load()); d < 300*time.Millisecond {
		t.Fatalf("ran after %v, before its delay", d)
	}
}

// [start shutdown]

func TestShutdownWaitsForRunningJobs(t *testing.T) {
	c := newClient(t)
	c.Enqueue(t.Context(), "q", "slow", EnqueueOptions{})
	started := make(chan struct{})
	stop := start(t, &Worker{Client: c, Queue: "q", Handler: func(ctx context.Context, j Job) error {
		close(started)
		time.Sleep(300 * time.Millisecond)
		return nil
	}})

	<-started
	begin := time.Now()
	stop() // cancel while the job is running
	t.Logf("Run returned %v after cancel", time.Since(begin).Round(10*time.Millisecond))

	counts, _ := c.Counts(t.Context(), "q")
	if counts["done"] != 1 {
		t.Fatalf("counts = %v: the running job should have finished", counts)
	}
}

// [end shutdown]
