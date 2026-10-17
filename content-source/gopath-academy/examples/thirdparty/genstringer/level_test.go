package genstringer

import "fmt"

func Example() {
	fmt.Println(LevelDebug, LevelWarn, Level(9))
	// Output: Debug Warn Level(9)
}
