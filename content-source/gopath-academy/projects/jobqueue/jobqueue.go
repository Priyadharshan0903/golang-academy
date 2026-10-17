// Package jobqueue is a background job queue stored in PostgreSQL.
//
// Producers enqueue jobs, possibly inside their own database
// transaction. Workers claim jobs with SELECT ... FOR UPDATE SKIP
// LOCKED, run a handler, and mark each job done, or schedule a retry
// with backoff, or give up after MaxAttempts. A job whose worker
// crashed is picked up again once its lock expires.
//
// Delivery is at least once: handlers must be idempotent.
package jobqueue

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

// Job is one unit of work.
type Job struct {
	ID          int64
	Queue       string
	Payload     json.RawMessage
	Attempts    int // including the current one
	MaxAttempts int
	LastError   string
}

// Client enqueues jobs and inspects the queue.
type Client struct {
	pool *pgxpool.Pool
}

// New returns a Client. Call Migrate once before use.
func New(pool *pgxpool.Pool) *Client { return &Client{pool: pool} }

// Migrate creates the jobs table if it does not exist.
func (c *Client) Migrate(ctx context.Context) error {
	_, err := c.pool.Exec(ctx, schema)
	return err
}

// [start enqueue]

// EnqueueOptions control a single job.
type EnqueueOptions struct {
	Delay       time.Duration // run no earlier than this from now
	MaxAttempts int           // default 5
}

// Executor is satisfied by *pgxpool.Pool, *pgx.Conn and pgx.Tx.
type Executor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Enqueue adds a job and returns its ID.
func (c *Client) Enqueue(ctx context.Context, queue string, payload any, opts EnqueueOptions) (int64, error) {
	return EnqueueWith(ctx, c.pool, queue, payload, opts)
}

// EnqueueWith adds a job using db, which can be a transaction: the job
// then exists only if the transaction commits. "Create the order and
// enqueue the confirmation email" becomes one atomic step, the main
// advantage of keeping a queue in the database you already use.
func EnqueueWith(ctx context.Context, db Executor, queue string, payload any, opts EnqueueOptions) (int64, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("encode payload: %w", err)
	}
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 5
	}
	var id int64
	err = db.QueryRow(ctx,
		`INSERT INTO jobs (queue, payload, max_attempts, run_at)
		 VALUES ($1, $2, $3, now() + make_interval(secs => $4))
		 RETURNING id`,
		queue, b, opts.MaxAttempts, opts.Delay.Seconds()).Scan(&id)
	return id, err
}

// [end enqueue]

// Counts returns the number of jobs in each status for a queue.
func (c *Client) Counts(ctx context.Context, queue string) (map[string]int, error) {
	rows, err := c.pool.Query(ctx, `SELECT status, count(*) FROM jobs WHERE queue = $1 GROUP BY status`, queue)
	if err != nil {
		return nil, err
	}
	out := map[string]int{}
	var status string
	var n int
	_, err = pgx.ForEachRow(rows, []any{&status, &n}, func() error {
		out[status] = n
		return nil
	})
	return out, err
}

// [start claim]

// claim locks the next runnable job in queue for visibility, or returns
// pgx.ErrNoRows if there is none. Runnable means queued and due, or
// running with an expired lock (its worker died).
//
// SKIP LOCKED is what makes this a queue: concurrent workers skip rows
// another transaction has locked, instead of waiting for them, so each
// job is claimed by exactly one worker.
func (c *Client) claim(ctx context.Context, queue string, visibility time.Duration) (Job, error) {
	var j Job
	var lastErr *string
	err := c.pool.QueryRow(ctx, `
		UPDATE jobs
		   SET status = 'running', attempts = attempts + 1,
		       locked_until = now() + make_interval(secs => $2), updated_at = now()
		 WHERE id = (
		       SELECT id FROM jobs
		        WHERE queue = $1
		          AND (status = 'queued' AND run_at <= now()
		               OR status = 'running' AND locked_until < now())
		        ORDER BY run_at, id
		        LIMIT 1
		          FOR UPDATE SKIP LOCKED)
		RETURNING id, queue, payload, attempts, max_attempts, last_error`,
		queue, visibility.Seconds(),
	).Scan(&j.ID, &j.Queue, &j.Payload, &j.Attempts, &j.MaxAttempts, &lastErr)
	if lastErr != nil {
		j.LastError = *lastErr
	}
	return j, err
}

// [end claim]

// [start finish]

// complete marks a job done. The attempts check makes sure we still own
// the job: if our lock expired and another worker claimed it, attempts
// has changed and this update does nothing.
func (c *Client) complete(ctx context.Context, j Job) error {
	_, err := c.pool.Exec(ctx,
		`UPDATE jobs SET status = 'done', locked_until = NULL, updated_at = now()
		  WHERE id = $1 AND attempts = $2`, j.ID, j.Attempts)
	return err
}

// fail records a failed attempt: retry after backoff, or give up.
func (c *Client) fail(ctx context.Context, j Job, cause error, backoff time.Duration) error {
	status := "queued"
	if j.Attempts >= j.MaxAttempts {
		status = "failed"
	}
	_, err := c.pool.Exec(ctx,
		`UPDATE jobs SET status = $3, last_error = $4, locked_until = NULL,
		        run_at = now() + make_interval(secs => $5), updated_at = now()
		  WHERE id = $1 AND attempts = $2`,
		j.ID, j.Attempts, status, cause.Error(), backoff.Seconds())
	return err
}

// [end finish]
