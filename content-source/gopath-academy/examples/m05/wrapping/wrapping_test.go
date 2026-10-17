package wrapping

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// [start wrap]

var ErrNotFound = errors.New("not found")

func findUser(id int) error {
	return fmt.Errorf("find user %d: %w", id, ErrNotFound)
}

func loadProfile(id int) error {
	if err := findUser(id); err != nil {
		return fmt.Errorf("load profile: %w", err)
	}
	return nil
}

func Example_wrapping() {
	err := loadProfile(42)
	fmt.Println(err)

	fmt.Println(err == ErrNotFound)          // false: err is a wrapper
	fmt.Println(errors.Is(err, ErrNotFound)) // true: Is walks the chain

	fmt.Println(errors.Unwrap(err)) // one layer down

	// Output:
	// load profile: find user 42: not found
	// false
	// true
	// find user 42: not found
}

// [end wrap]

// [start verbs]

func Example_wrapVsFormat() {
	base := errors.New("disk full")

	wrapped := fmt.Errorf("save: %w", base)   // keeps the chain
	flattened := fmt.Errorf("save: %v", base) // only copies the text

	fmt.Println(wrapped, "|", errors.Is(wrapped, base))
	fmt.Println(flattened, "|", errors.Is(flattened, base))

	// Output:
	// save: disk full | true
	// save: disk full | false
}

// [end verbs]

// [start as]

func Example_errorsAs() {
	_, err := os.Open("/definitely/not/here.txt")
	err = fmt.Errorf("read config: %w", err)

	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		fmt.Println("op:", pathErr.Op)
		fmt.Println("path:", pathErr.Path)
	}
	fmt.Println(errors.Is(err, fs.ErrNotExist))

	// Output:
	// op: open
	// path: /definitely/not/here.txt
	// true
}

// [end as]

// [start custom]

// QueryError wraps a lower-level error and adds the query.
type QueryError struct {
	Query string
	Err   error
}

func (e *QueryError) Error() string { return e.Query + ": " + e.Err.Error() }

// Unwrap lets errors.Is and errors.As see the wrapped error.
func (e *QueryError) Unwrap() error { return e.Err }

var ErrTimeout = errors.New("timeout")

func Example_customWrapper() {
	err := fmt.Errorf("list users: %w", &QueryError{Query: "SELECT * FROM users", Err: ErrTimeout})

	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrTimeout))

	var qe *QueryError
	fmt.Println(errors.As(err, &qe), qe.Query)

	// Output:
	// list users: SELECT * FROM users: timeout
	// true
	// true SELECT * FROM users
}

// [end custom]
