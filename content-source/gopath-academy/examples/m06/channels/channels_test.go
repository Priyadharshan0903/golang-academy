package channels

import (
	"fmt"
	"sync"
)

// [start unbuffered]

func Example_unbuffered() {
	ch := make(chan string) // unbuffered: a send waits for a receiver

	go func() {
		ch <- "ping" // blocks until main is ready to receive
	}()

	msg := <-ch // blocks until the goroutine sends
	fmt.Println(msg)

	// Output:
	// ping
}

// [end unbuffered]

// [start buffered]

func Example_buffered() {
	ch := make(chan int, 2) // room for 2 values without a receiver

	ch <- 1 // does not block
	ch <- 2 // does not block
	fmt.Println(len(ch), cap(ch))
	// A third send would block until someone receives.

	fmt.Println(<-ch, <-ch)

	// Output:
	// 2 2
	// 1 2
}

// [end buffered]

// [start close]

func produce(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // the sender closes, telling receivers "no more values"
		for i := range n {
			out <- i * i
		}
	}()
	return out
}

func Example_closeAndRange() {
	for v := range produce(4) { // range stops when the channel is closed
		fmt.Println(v)
	}

	ch := make(chan int, 1)
	ch <- 7
	close(ch)
	v, ok := <-ch
	fmt.Println(v, ok) // buffered values are still delivered after close
	v, ok = <-ch
	fmt.Println(v, ok) // then receives return the zero value and false

	// Output:
	// 0
	// 1
	// 4
	// 9
	// 7 true
	// 0 false
}

// [end close]

// [start direction]

// Direction types document and enforce who sends and who receives.
func sendOnly(out chan<- string)       { out <- "hello" }
func recvOnly(in <-chan string) string { return <-in }

func Example_directions() {
	ch := make(chan string, 1) // bidirectional
	sendOnly(ch)               // converts to chan<- string automatically
	fmt.Println(recvOnly(ch))  // converts to <-chan string
	// Inside sendOnly, <-out would not compile.

	// Output:
	// hello
}

// [end direction]

// [start signal]

func Example_doneSignal() {
	done := make(chan struct{}) // struct{}: the event matters, not a value
	var wg sync.WaitGroup

	for i := range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-done // every waiter is released by one close
			_ = i
		}()
	}
	close(done) // broadcast: closing wakes ALL receivers
	wg.Wait()
	fmt.Println("all released")

	// Output:
	// all released
}

// [end signal]
