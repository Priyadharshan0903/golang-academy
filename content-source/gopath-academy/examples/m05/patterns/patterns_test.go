package patterns

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// [start mapping]

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("already exists")
	ErrValidation = errors.New("validation failed")
)

// statusFor translates domain errors into HTTP status codes in ONE
// place, at the edge of the program.
func statusFor(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func Example_mapAtTheEdge() {
	fmt.Println(statusFor(nil))
	fmt.Println(statusFor(fmt.Errorf("get order 7: %w", ErrNotFound)))
	fmt.Println(statusFor(fmt.Errorf("create user: %w", ErrConflict)))
	fmt.Println(statusFor(errors.New("database on fire")))

	// Output:
	// 200
	// 404
	// 409
	// 500
}

// [end mapping]

// [start retry]

var errTransient = errors.New("temporarily unavailable")

// retry calls f up to attempts times, backing off between tries, and
// gives up immediately on errors that retrying cannot fix.
func retry(attempts int, base time.Duration, f func() error) error {
	var err error
	for i := range attempts {
		err = f()
		if err == nil || !errors.Is(err, errTransient) {
			return err
		}
		time.Sleep(base << i) // 1x, 2x, 4x, ...
	}
	return fmt.Errorf("after %d attempts: %w", attempts, err)
}

func Example_retry() {
	calls := 0
	err := retry(5, time.Millisecond, func() error {
		calls++
		if calls < 3 {
			return errTransient
		}
		return nil
	})
	fmt.Println(err, calls)

	err = retry(3, time.Millisecond, func() error { return errTransient })
	fmt.Println(err)

	err = retry(3, time.Millisecond, func() error { return ErrValidation })
	fmt.Println(err)

	// Output:
	// <nil> 3
	// after 3 attempts: temporarily unavailable
	// validation failed
}

// [end retry]

// [start errwriter]

// errWriter remembers the first error so a sequence of writes can be
// checked once at the end (from Rob Pike's "Errors are values").
type errWriter struct {
	out []string
	err error
}

func (w *errWriter) write(s string) {
	if w.err != nil {
		return // already failed: skip the rest
	}
	if s == "" {
		w.err = errors.New("empty line")
		return
	}
	w.out = append(w.out, s)
}

func Example_errWriter() {
	w := &errWriter{}
	w.write("header")
	w.write("")
	w.write("body") // skipped
	fmt.Println(w.out, w.err)

	// Output:
	// [header] empty line
}

// [end errwriter]
