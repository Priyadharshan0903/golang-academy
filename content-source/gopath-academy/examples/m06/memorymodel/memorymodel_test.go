package memorymodel

import (
	"fmt"
	"sync"
)

// [start handoff]

type result struct {
	values []int
	sum    int
}

func Example_channelHandoff() {
	ch := make(chan *result)

	go func() {
		r := &result{values: []int{1, 2, 3}}
		for _, v := range r.values {
			r.sum += v
		}
		ch <- r // the send "happens before" the receive completes...
		// ...so after this point the goroutine must not touch r again.
	}()

	r := <-ch // ...and everything written before the send is visible here
	fmt.Println(r.sum)

	// Output:
	// 6
}

// [end handoff]

// [start broken]

// The spin flag below has NO synchronization. The memory model gives no
// guarantee that the loop ever sees ready == true, or that it sees msg.
// It is shown in a comment only; never write this:
//
//	var msg string
//	var ready bool
//	go func() { msg = "hi"; ready = true }()
//	for !ready {}    // may loop forever
//	fmt.Println(msg) // may print ""
//
// The fix is any synchronizing operation: a channel, a mutex, a
// WaitGroup or an atomic.

func Example_fixedWithWaitGroup() {
	var msg string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg = "hi"
	}()
	wg.Wait() // Done happens before Wait returns, so msg is visible
	fmt.Println(msg)

	// Output:
	// hi
}

// [end broken]
