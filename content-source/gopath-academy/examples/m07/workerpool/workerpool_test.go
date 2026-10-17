package workerpool

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// [start pool]

type job struct {
	id   int
	text string
}

type result struct {
	id    int
	words int
}

// countWords runs a fixed number of workers over a stream of jobs.
func countWords(ctx context.Context, texts []string, workers int) []int {
	jobs := make(chan job)
	results := make(chan result)

	// 1. Start the workers. Each reads jobs until the channel closes.
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := result{id: j.id, words: len(strings.Fields(j.text))}
				select {
				case results <- r:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// 2. Feed the jobs, then close the channel so workers stop.
	go func() {
		defer close(jobs)
		for i, t := range texts {
			select {
			case jobs <- job{id: i, text: t}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 3. Close results once every worker is done.
	go func() {
		wg.Wait()
		close(results)
	}()

	// 4. Collect. Results arrive in any order, so place them by id.
	counts := make([]int, len(texts))
	for r := range results {
		counts[r.id] = r.words
	}
	return counts
}

func Example_workerPool() {
	texts := []string{
		"go is fun",
		"channels connect goroutines",
		"",
		"a worker pool bounds concurrency to a fixed size",
	}
	fmt.Println(countWords(context.Background(), texts, 2))

	// Output:
	// [3 3 0 9]
}

// [end pool]
