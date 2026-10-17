//go:build !solution

// Package pmap runs a function over a slice in parallel, with a
// concurrency limit, results in input order, and cancellation on the
// first error. It combines worker pools, context and generics.
package pmap

import "context"

// Map calls f for every input using at most workers goroutines and
// returns the results in the same order as inputs.
//
// If any call returns an error, Map cancels the context passed to the
// other calls, waits for every goroutine it started to finish, and
// returns the first error. If ctx itself is cancelled, Map returns an
// error wrapping ctx.Err().
//
// workers is at least 1. Map must not leak goroutines.
func Map[T, R any](ctx context.Context, inputs []T, workers int, f func(context.Context, T) (R, error)) ([]R, error) {
	// TODO: implement me.
	return nil, nil
}
