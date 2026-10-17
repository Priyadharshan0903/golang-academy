package jobqueue

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// A Handler processes one job. Returning an error schedules a retry.
// The context is cancelled when the job's lock is about to expire.
type Handler func(ctx context.Context, j Job) error

// Worker runs a handler for the jobs of one queue.
type Worker struct {
	Client       *Client
	Queue        string
	Handler      Handler
	Concurrency  int                             // goroutines processing jobs, default 4
	PollInterval time.Duration                   // wait when the queue is empty, default 1s
	Visibility   time.Duration                   // how long a claimed job stays locked, default 1m
	Backoff      func(attempt int) time.Duration // delay before retry n, default exponential
	Logger       *slog.Logger
}

// [start run]

// Run processes jobs until ctx is cancelled. It then stops claiming new
// jobs, waits for the jobs in progress to finish, and returns.
func (w *Worker) Run(ctx context.Context) error {
	w.defaults()
	var wg sync.WaitGroup
	for range w.Concurrency {
		wg.Go(func() { w.loop(ctx) })
	}
	wg.Wait()
	return nil
}

func (w *Worker) loop(ctx context.Context) {
	for ctx.Err() == nil {
		// The claim itself is not cancelled by shutdown: a claim that
		// commits after we gave up waiting would leave a job locked
		// until its visibility timeout.
		claimCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		j, err := w.Client.claim(claimCtx, w.Queue, w.Visibility)
		cancel()
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			// Nothing to do: wait, with jitter so idle workers do not
			// poll in lockstep.
			sleep(ctx, w.PollInterval/2+rand.N(w.PollInterval))
		case err != nil:
			w.Logger.Error("claim failed", "queue", w.Queue, "error", err)
			sleep(ctx, w.PollInterval)
		default:
			w.process(j)
		}
	}
}

// process runs one job. It deliberately does not use the Run context:
// shutting down must let a running job finish. Its own deadline is the
// end of its lock, after which another worker may claim it.
func (w *Worker) process(j Job) {
	ctx, cancel := context.WithTimeout(context.Background(), w.Visibility)
	defer cancel()

	if j.Attempts > j.MaxAttempts { // a worker crashed during the last allowed attempt
		w.finish(j, w.Client.fail(ctx, j, errors.New("worker stopped during the final attempt"), 0))
		return
	}
	start := time.Now()
	err := safeCall(ctx, w.Handler, j)
	if err == nil {
		w.finish(j, w.Client.complete(ctx, j))
		w.Logger.Info("job done", "id", j.ID, "attempt", j.Attempts, "duration", time.Since(start))
		return
	}
	w.Logger.Warn("job failed", "id", j.ID, "attempt", j.Attempts, "max_attempts", j.MaxAttempts, "error", err)
	w.finish(j, w.Client.fail(ctx, j, err, w.Backoff(j.Attempts)))
}

// [end run]

func (w *Worker) finish(j Job, err error) {
	if err != nil {
		w.Logger.Error("could not record job result", "id", j.ID, "error", err)
	}
}

// safeCall runs the handler and turns a panic into an error, so one bad
// job cannot crash the worker process.
func safeCall(ctx context.Context, h Handler, j Job) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("panic: %v", v)
		}
	}()
	return h(ctx, j)
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

func (w *Worker) defaults() {
	if w.Concurrency <= 0 {
		w.Concurrency = 4
	}
	if w.PollInterval <= 0 {
		w.PollInterval = time.Second
	}
	if w.Visibility <= 0 {
		w.Visibility = time.Minute
	}
	if w.Backoff == nil {
		w.Backoff = ExponentialBackoff(time.Second, time.Hour)
	}
	if w.Logger == nil {
		w.Logger = slog.Default()
	}
}

// ExponentialBackoff doubles the delay after each attempt, from base up
// to limit, with up to 20% random jitter.
func ExponentialBackoff(base, limit time.Duration) func(int) time.Duration {
	return func(attempt int) time.Duration {
		d := base << min(attempt-1, 30)
		if d <= 0 || d > limit {
			d = limit
		}
		return d + rand.N(d/5+1)
	}
}
