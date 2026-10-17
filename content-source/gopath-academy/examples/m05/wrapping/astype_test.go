//go:build go1.26

package wrapping

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// [start astype]

func Example_asType() {
	_, err := os.Open("/definitely/not/here.txt")
	err = fmt.Errorf("read config: %w", err)

	// errors.AsType (Go 1.26+) is a generic, type-safe form of errors.As:
	// no target variable, no pointer-to-pointer.
	if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
		fmt.Println("path:", pathErr.Path)
	}

	// Output:
	// path: /definitely/not/here.txt
}

// [end astype]
