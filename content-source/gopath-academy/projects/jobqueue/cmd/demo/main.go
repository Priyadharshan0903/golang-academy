// Command demo enqueues a batch of fake email jobs and processes them
// with a pool of workers. One in five sends "fails" the first time, to
// show retries. Run it against any PostgreSQL database:
//
//	DATABASE_URL='postgres://postgres:pg@localhost:5432/postgres?sslmode=disable' go run ./cmd/demo
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/jobqueue"
	"github.com/jackc/pgx/v5/pgxpool"
)

type email struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demo:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()
	client := jobqueue.New(pool)
	if err := client.Migrate(ctx); err != nil {
		return err
	}

	for i := range 20 {
		_, err := client.Enqueue(ctx, "emails", email{To: fmt.Sprintf("user%d@example.com", i), Subject: "Welcome"}, jobqueue.EnqueueOptions{})
		if err != nil {
			return err
		}
	}

	w := &jobqueue.Worker{
		Client:      client,
		Queue:       "emails",
		Concurrency: 4,
		Backoff:     jobqueue.ExponentialBackoff(200*time.Millisecond, 5*time.Second),
		Logger:      logger,
		Handler: func(ctx context.Context, j jobqueue.Job) error {
			var e email
			if err := json.Unmarshal(j.Payload, &e); err != nil {
				return err
			}
			time.Sleep(100 * time.Millisecond) // "send" the email
			if j.ID%5 == 0 && j.Attempts == 1 {
				return errors.New("smtp: connection reset")
			}
			return nil
		},
	}

	workCtx, cancel := context.WithCancel(ctx)
	go func() {
		for workCtx.Err() == nil {
			time.Sleep(200 * time.Millisecond)
			if c, _ := client.Counts(ctx, "emails"); c["queued"] == 0 && c["running"] == 0 {
				logger.Info("queue empty", "counts", c)
				cancel()
			}
		}
	}()
	return w.Run(workCtx)
}
