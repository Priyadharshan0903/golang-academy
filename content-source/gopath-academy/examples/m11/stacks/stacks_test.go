package stacks

import "testing"

func TestDeepRecursion(t *testing.T) {
	done := make(chan int)
	go func() { done <- Sum(1_000_000) }() // a fresh goroutine starts with a small stack
	if got := <-done; got != 500000500000 {
		t.Fatalf("Sum = %d", got)
	}
}

func TestStackMoves(t *testing.T) {
	done := make(chan [2]uintptr)
	go func() {
		b, a := StackMoves()
		done <- [2]uintptr{b, a}
	}()
	r := <-done
	t.Logf("before %#x, after %#x", r[0], r[1])
	if r[0] == r[1] {
		t.Fatal("expected the stack to move when it grew")
	}
}

func TestStackCost(t *testing.T) {
	per := StackBytesPerGoroutine(10_000)
	t.Logf("about %d bytes of stack per blocked goroutine", per)
	if per > 16<<10 {
		t.Fatalf("goroutines should start with small stacks, got %d bytes", per)
	}
}
