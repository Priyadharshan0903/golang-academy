package kinds

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// [start sentinel]

// Sentinel errors: exported variables callers compare with errors.Is.
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
)

func Example_sentinel() {
	r := strings.NewReader("")
	_, err := r.Read(make([]byte, 1))
	fmt.Println(errors.Is(err, io.EOF)) // io.EOF is the most famous sentinel

	err = fmt.Errorf("get /admin: %w", ErrUnauthorized)
	switch {
	case errors.Is(err, ErrNotFound):
		fmt.Println("404")
	case errors.Is(err, ErrUnauthorized):
		fmt.Println("401")
	}

	// Output:
	// true
	// 401
}

// [end sentinel]

// [start typed]

// RateLimitError is an error type: callers need its data, not just its identity.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %s", e.RetryAfter)
}

func callAPI() error { return &RateLimitError{RetryAfter: 30 * time.Second} }

func Example_errorType() {
	err := fmt.Errorf("sync orders: %w", callAPI())

	var rl *RateLimitError
	if errors.As(err, &rl) {
		fmt.Println("waiting", rl.RetryAfter)
	}

	// Output:
	// waiting 30s
}

// [end typed]

// [start behavior]

// Opaque errors: callers check for a behavior through a small
// interface instead of depending on a concrete type or variable.
type temporary interface {
	Temporary() bool
}

type netGlitch struct{}

func (netGlitch) Error() string   { return "connection reset" }
func (netGlitch) Temporary() bool { return true }

func isTemporary(err error) bool {
	var t temporary
	return errors.As(err, &t) && t.Temporary()
}

func Example_behavior() {
	fmt.Println(isTemporary(fmt.Errorf("fetch: %w", netGlitch{})))
	fmt.Println(isTemporary(errors.New("bad request")))

	// Output:
	// true
	// false
}

// [end behavior]

// [start is]

// A custom Is method lets one error match a whole category.
type HTTPError struct{ Code int }

func (e *HTTPError) Error() string { return fmt.Sprintf("http %d", e.Code) }

var ErrServer = errors.New("server error")

func (e *HTTPError) Is(target error) bool {
	return target == ErrServer && e.Code >= 500
}

func Example_customIs() {
	fmt.Println(errors.Is(&HTTPError{Code: 503}, ErrServer))
	fmt.Println(errors.Is(&HTTPError{Code: 404}, ErrServer))

	// Output:
	// true
	// false
}

// [end is]
