package color

import (
	"encoding/json"
	"fmt"
)

// [start example]

func Example() {
	fmt.Println(ColorAmber, Color(42))

	c, err := ParseColor("green")
	fmt.Println(c == ColorGreen, err)
	_, err = ParseColor("purple")
	fmt.Println(err)

	b, _ := json.Marshal(map[string]Color{"light": ColorRed})
	fmt.Println(string(b), ColorValues())
	// Output:
	// amber Color(42)
	// true <nil>
	// invalid Color "purple"
	// {"light":"red"} [red amber green]
}

// [end example]
