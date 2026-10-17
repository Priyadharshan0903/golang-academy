//go:build !solution

// Package pipeline practices channels: generators, stages that read
// from one channel and write to another, and merging channels.
//
// Rules for every function here: the function that creates a channel
// is the one that closes it, and every goroutine you start must exit
// once its input is exhausted.
package pipeline

// Generate returns a channel that yields nums in order, then closes.
func Generate(nums ...int) <-chan int {
	// TODO: implement me.
	out := make(chan int)
	close(out)
	return out
}

// Square reads every value from in and sends its square on the
// returned channel, which is closed after in is closed and drained.
func Square(in <-chan int) <-chan int {
	// TODO: implement me.
	out := make(chan int)
	close(out)
	return out
}

// Merge sends every value from all input channels on one output
// channel (in any order) and closes it once ALL inputs are closed.
//
// Hint: one goroutine per input, plus a sync.WaitGroup and one more
// goroutine that closes the output after Wait returns.
func Merge(chans ...<-chan int) <-chan int {
	// TODO: implement me.
	out := make(chan int)
	close(out)
	return out
}
