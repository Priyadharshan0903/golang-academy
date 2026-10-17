package contexts

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// [start cancel]

// work does steps until it finishes or ctx is cancelled.
func work(ctx context.Context, steps int) (done int, err error) {
	for i := range steps {
		select {
		case <-ctx.Done(): // closed when the context is cancelled
			return i, ctx.Err()
		case <-time.After(5 * time.Millisecond): // one unit of work
		}
	}
	return steps, nil
}

func Example_cancel() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(12 * time.Millisecond)
		cancel() // tell everything using ctx to stop
	}()

	done, err := work(ctx, 100)
	fmt.Println(done < 100, err)
	fmt.Println(errors.Is(err, context.Canceled))

	// Output:
	// true context canceled
	// true
}

// [end cancel]

// [start timeout]

func Example_timeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel() // ALWAYS call cancel to release the timer early

	_, err := work(ctx, 100)
	fmt.Println(err)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))

	deadline, ok := ctx.Deadline()
	fmt.Println(ok, !deadline.IsZero())

	// Output:
	// context deadline exceeded
	// true
	// true true
}

// [end timeout]

// [start tree]

func Example_cancellationTree() {
	parent, cancelParent := context.WithCancel(context.Background())
	child, cancelChild := context.WithTimeout(parent, time.Hour)
	defer cancelChild()

	cancelParent() // cancelling a parent cancels every child

	<-child.Done()
	fmt.Println(child.Err())

	// Output:
	// context canceled
}

// [end tree]

// [start cause]

var errShuttingDown = errors.New("server shutting down")

func Example_cancelCause() {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(errShuttingDown) // record WHY it was cancelled

	fmt.Println(ctx.Err())          // still the standard error
	fmt.Println(context.Cause(ctx)) // the reason you gave

	// Output:
	// context canceled
	// server shutting down
}

// [end cause]

// [start values]

// An unexported key type prevents collisions with other packages.
type ctxKey int

const requestIDKey ctxKey = iota

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

func Example_values() {
	ctx := WithRequestID(context.Background(), "req-7f3a")
	id, ok := RequestID(ctx)
	fmt.Println(id, ok)

	_, ok = RequestID(context.Background())
	fmt.Println(ok)

	// Output:
	// req-7f3a true
	// false
}

// [end values]

// [start afterfunc]

func Example_afterFunc() {
	ctx, cancel := context.WithCancel(context.Background())
	stopped := make(chan struct{})

	// AfterFunc (Go 1.21+) runs f in its own goroutine once ctx is done.
	context.AfterFunc(ctx, func() {
		fmt.Println("cleanup after cancel")
		close(stopped)
	})

	cancel()
	<-stopped

	// WithoutCancel keeps values but ignores the parent's cancellation,
	// for work that must finish, like writing an audit log.
	detached := context.WithoutCancel(ctx)
	fmt.Println(ctx.Err(), detached.Err())

	// Output:
	// cleanup after cancel
	// context canceled <nil>
}

// [end afterfunc]
