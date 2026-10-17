// Package errgroupdemo shows golang.org/x/sync/errgroup (lesson 7.4).
package errgroupdemo

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// [start errgroup]

// fetchAll fetches every URL with at most 3 requests in flight, and
// stops all of them as soon as one fails.
func fetchAll(ctx context.Context, urls []string, fetch func(context.Context, string) (string, error)) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx) // ctx is cancelled on the first error
	g.SetLimit(3)                       // bounded concurrency

	var mu sync.Mutex
	var bodies []string
	for _, u := range urls {
		g.Go(func() error { // blocks while 3 are already running
			body, err := fetch(ctx, u)
			if err != nil {
				return fmt.Errorf("fetch %s: %w", u, err)
			}
			mu.Lock()
			bodies = append(bodies, body)
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil { // waits for all, returns the first error
		return nil, err
	}
	sort.Strings(bodies)
	return bodies, nil
}

// [end errgroup]

func fakeFetch(ctx context.Context, u string) (string, error) {
	select {
	case <-time.After(time.Millisecond):
	case <-ctx.Done():
		return "", ctx.Err()
	}
	if u == "bad" {
		return "", fmt.Errorf("status 500")
	}
	return "body of " + u, nil
}

func Example_fetchAll() {
	bodies, err := fetchAll(context.Background(), []string{"b", "a", "c", "d"}, fakeFetch)
	fmt.Println(bodies, err)

	_, err = fetchAll(context.Background(), []string{"a", "bad", "c"}, fakeFetch)
	fmt.Println(err)

	// Output:
	// [body of a body of b body of c body of d] <nil>
	// fetch bad: status 500
}
