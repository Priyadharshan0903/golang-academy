// Command gcdemo runs gc.Churn once with the GOGC and GOMEMLIMIT
// settings from the environment:
//
//	GOGC=50 go run ./examples/m11/gc/gcdemo
//	GOGC=off GOMEMLIMIT=64MiB go run ./examples/m11/gc/gcdemo
package main

import (
	"fmt"
	"os"

	"github.com/Priyadharshan0903/gopath-academy/examples/m11/gc"
)

func main() {
	r := gc.Churn(16<<20, 512<<20) // 16 MiB live, 512 MiB of garbage
	fmt.Printf("GOGC=%s GOMEMLIMIT=%s: %d GC cycles, heap goal about %d MiB\n",
		or(os.Getenv("GOGC"), "100"), or(os.Getenv("GOMEMLIMIT"), "none"), r.Cycles, r.MedianGoal>>20)
}

func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
