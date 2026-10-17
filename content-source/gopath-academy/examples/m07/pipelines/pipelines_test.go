package pipelines

import (
	"context"
	"fmt"
	"slices"
	"sync"
)

// [start stages]

// Each stage takes ctx so the whole pipeline can be stopped early.

func generate(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

func square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// merge fans in several channels (lesson 6.2), also cancellable.
func merge(ctx context.Context, chans ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	for _, c := range chans {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range c {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	go func() { wg.Wait(); close(out) }()
	return out
}

// [end stages]

// [start fanout]

func Example_fanOutFanIn() {
	ctx := context.Background()
	src := generate(ctx, 1, 2, 3, 4, 5, 6)

	// Fan out: three square stages read from the SAME source channel,
	// so the work is shared between them.
	workers := []<-chan int{square(ctx, src), square(ctx, src), square(ctx, src)}

	// Fan in: merge their outputs into one channel.
	var got []int
	for v := range merge(ctx, workers...) {
		got = append(got, v)
	}
	slices.Sort(got)
	fmt.Println(got)

	// Output:
	// [1 4 9 16 25 36]
}

// [end fanout]

// [start early]

func Example_stopEarly() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // stops every stage when we return

	results := square(ctx, generate(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
	for v := range results {
		fmt.Println(v)
		if v >= 9 {
			break // the deferred cancel lets the stages exit instead of leaking
		}
	}

	// Output:
	// 1
	// 4
	// 9
}

// [end early]
