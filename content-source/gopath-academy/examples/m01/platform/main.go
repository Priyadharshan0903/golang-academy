// Command platform prints the operating system and CPU architecture
// the binary was compiled for. Build it with different GOOS and GOARCH
// values to see cross-compilation in action.
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("built for %s/%s with %s\n", runtime.GOOS, runtime.GOARCH, runtime.Version())
}
