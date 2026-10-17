package selects

import (
	"fmt"
	"time"
)

// [start select]

func Example_select() {
	fast := make(chan string, 1)
	slow := make(chan string, 1)
	fast <- "fast result"

	// select waits on several channel operations and runs the first
	// one that can proceed.
	select {
	case msg := <-fast:
		fmt.Println(msg)
	case msg := <-slow:
		fmt.Println(msg)
	}

	// Output:
	// fast result
}

// [end select]

// [start timeout]

func slowLookup() <-chan string {
	ch := make(chan string, 1) // buffered: the goroutine can finish even if nobody reads
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch <- "result"
	}()
	return ch
}

func Example_timeout() {
	select {
	case r := <-slowLookup():
		fmt.Println(r)
	case <-time.After(20 * time.Millisecond):
		fmt.Println("timed out")
	}

	// Output:
	// timed out
}

// [end timeout]

// [start default]

func trySend(ch chan<- int, v int) bool {
	select {
	case ch <- v:
		return true
	default: // runs immediately if no other case is ready
		return false
	}
}

func Example_nonBlocking() {
	queue := make(chan int, 1)
	fmt.Println(trySend(queue, 1)) // room in the buffer
	fmt.Println(trySend(queue, 2)) // full: drop instead of blocking

	// Output:
	// true
	// false
}

// [end default]

// [start nilcase]

func Example_nilChannelDisablesCase() {
	a := make(chan int)
	b := make(chan int)
	go func() { a <- 1; close(a) }()
	go func() { b <- 2; close(b) }()

	sum := 0
	for a != nil || b != nil {
		select {
		case v, ok := <-a:
			if !ok {
				a = nil // a nil channel blocks forever, so this case is now ignored
				continue
			}
			sum += v
		case v, ok := <-b:
			if !ok {
				b = nil
				continue
			}
			sum += v
		}
	}
	fmt.Println(sum)

	// Output:
	// 3
}

// [end nilcase]

// [start ticker]

func Example_ticker() {
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop() // release the ticker when done

	stop := time.NewTimer(60 * time.Millisecond)
	defer stop.Stop()

	ticks := 0
	for {
		select {
		case <-ticker.C:
			ticks++
		case <-stop.C:
			fmt.Println(ticks >= 2) // timing varies, so only check a lower bound
			return
		}
	}

	// Output:
	// true
}

// [end ticker]
