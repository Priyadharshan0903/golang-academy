//go:build go1.27

package formatting

import (
	"fmt"
	"strings"
)

// [start cutlast]

func Example_cutLast() {
	// strings.CutLast (Go 1.27+) splits around the LAST separator.
	dir, file, ok := strings.CutLast("/var/log/app/server.log", "/")
	fmt.Println(dir, file, ok)

	host, port, _ := strings.CutLast("[::1]:8080", ":")
	fmt.Println(host, port)

	// Output:
	// /var/log/app server.log true
	// [::1] 8080
}

// [end cutlast]
