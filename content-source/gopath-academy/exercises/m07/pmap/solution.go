//go:build solution

// Package pmap runs a function over a slice in parallel, with a
// concurrency limit, results in input order, and cancellation on the
// first error. It combines worker pools, context and generics.
package pmap

import (
	"context"
	"sync"
)

// Map calls f for every input using at most workers goroutines.
func Map[T, R any](ctx context.Context, inputs []T, workers int, f func(context.Context, T) (R, error)) ([]R, error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	results := make([]R, len(inputs)) // each index written by one goroutine
	indexes := make(chan int)

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)
	fail := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel(err)
		})
	}

	for range max(1, workers) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range indexes {
				r, err := f(ctx, inputs[i])
				if err != nil {
					fail(err)
					return
				}
				results[i] = r
			}
		}()
	}

feed:
	for i := range inputs {
		select {
		case indexes <- i:
		case <-ctx.Done():
			break feed
		}
	}
	close(indexes)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
