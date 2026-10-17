package constants

import (
	"fmt"
	"time"
)

// [start untyped]

func Example_untyped() {
	const timeout = 5 // untyped constant: it has no fixed type yet

	var a int64 = timeout   // fine
	var b float64 = timeout // also fine
	d := timeout * time.Second

	fmt.Println(a, b, d)

	// Untyped constants are computed with very high precision.
	const huge = 1 << 100
	fmt.Println(huge >> 98)

	// Output:
	// 5 5 5s
	// 4
}

// [end untyped]

// [start iota]

// Weekday is a named integer type. Named types are covered in Module 3.
type Weekday int

const (
	Sunday  Weekday = iota // 0
	Monday                 // 1, the expression repeats with iota incremented
	Tuesday                // 2
	Wednesday
	Thursday
	Friday
	Saturday
)

func Example_iota() {
	fmt.Println(Sunday, Wednesday, Saturday)

	// Output:
	// 0 3 6
}

// [end iota]

// [start iotaexpr]

type ByteSize int64

const (
	_           = iota // skip 0 with the blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
)

type Permission uint8

const (
	Read    Permission = 1 << iota // 1
	Write                          // 2
	Execute                        // 4
)

func Example_iotaExpressions() {
	fmt.Println(KB, MB, GB)

	perms := Read | Write
	fmt.Println(perms, perms&Write != 0, perms&Execute != 0)

	// Output:
	// 1024 1048576 1073741824
	// 3 true false
}

// [end iotaexpr]
