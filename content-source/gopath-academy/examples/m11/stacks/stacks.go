// Package stacks holds the examples for lesson 11.2.
package stacks

import (
	"runtime"
	"unsafe"
)

// [start sum]

// Sum adds 1..n recursively. Go has no tail-call optimization, so this
// uses one stack frame per call: a million calls need tens of
// megabytes of stack, which the runtime grows on demand.
func Sum(n int) int {
	if n == 0 {
		return 0
	}
	return n + Sum(n-1)
}

// [end sum]

// [start moves]

// StackMoves reports the address of a local variable before and after
// a deep recursion. When the stack grows, the runtime copies it to a
// new, larger allocation, and every pointer into it is updated.
func StackMoves() (before, after uintptr) {
	var x int
	before = uintptr(unsafe.Pointer(&x)) // a number, not a pointer: x stays on the stack
	Sum(100_000)                         // forces the stack to grow
	after = uintptr(unsafe.Pointer(&x))
	return before, after
}

// [end moves]

// [start cost]

// StackBytesPerGoroutine starts n goroutines that block, and reports
// how much stack memory each one added on average.
func StackBytesPerGoroutine(n int) uint64 {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	release := make(chan struct{})
	started := make(chan struct{})
	for range n {
		go func() {
			started <- struct{}{}
			<-release
		}()
	}
	for range n {
		<-started
	}
	runtime.ReadMemStats(&after)
	close(release)

	return (after.StackInuse - before.StackInuse) / uint64(n)
}

// [end cost]
