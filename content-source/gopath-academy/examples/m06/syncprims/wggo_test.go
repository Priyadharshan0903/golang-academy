//go:build go1.25

package syncprims

import (
	"fmt"
	"sync"
)

// [start wggo]

func Example_waitGroupGo() {
	var wg sync.WaitGroup
	results := make([]int, 4)
	for i := range results {
		// WaitGroup.Go (Go 1.25+) calls Add(1), starts the goroutine,
		// and calls Done when the function returns.
		wg.Go(func() {
			results[i] = i * 10
		})
	}
	wg.Wait()
	fmt.Println(results)

	// Output:
	// [0 10 20 30]
}

// [end wggo]
