// [start external]

// The _test suffix makes this an external test package. It can only
// use the exported API, exactly like real users of the package.
package greeting_test

import (
	"fmt"

	"github.com/Priyadharshan0903/gopath-academy/examples/m01/greeting"
)

func ExampleHello() {
	fmt.Println(greeting.Hello("Dharshan"))
	fmt.Println(greeting.Hello("   "))
	// greeting.normalize("x") would not compile here: it is unexported.

	// Output:
	// Hello, Dharshan!
	// Hello, Gopher!
}

// [end external]
