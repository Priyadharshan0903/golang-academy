package numbers

import (
	"fmt"
	"math"
	"strconv"
)

// [start conversion]

func Example_conversion() {
	var i int = 42
	var f float64 = float64(i) // conversions are always explicit
	var u uint8 = uint8(i)

	// Converting a float to an int truncates toward zero.
	// (int(3.99) on a constant would not even compile: the
	// compiler refuses constant conversions that lose information.)
	pos, neg := 3.99, -3.99
	fmt.Println(int(pos), int(neg))
	fmt.Println(f, u)

	// Integer division truncates too. Convert first for a float result.
	fmt.Println(7/2, float64(7)/2)

	// Output:
	// 3 -3
	// 42 42
	// 3 3.5
}

// [end conversion]

// [start overflow]

func Example_overflow() {
	var small int8 = 127
	small++ // wraps around silently at run time
	fmt.Println(small)

	var count uint = 0
	count-- // unsigned underflow wraps to the maximum value
	fmt.Println(count == math.MaxUint)

	// A conversion that does not fit keeps only the low bits.
	big := 300
	fmt.Println(uint8(big))

	// Output:
	// -128
	// true
	// 44
}

// [end overflow]

// [start floats]

func Example_floats() {
	a, b := 0.1, 0.2
	fmt.Println(a+b == 0.3)
	fmt.Println(a + b)

	// Compare floats with a tolerance instead.
	fmt.Println(math.Abs((a+b)-0.3) < 1e-9)

	// Output:
	// false
	// 0.30000000000000004
	// true
}

// [end floats]

// [start strconv]

func Example_strconv() {
	n, err := strconv.Atoi("123")
	fmt.Println(n+1, err)

	_, err = strconv.Atoi("12a")
	fmt.Println(err)

	// string(65) is NOT "65": it converts a number to a character.
	fmt.Println(strconv.Itoa(65), string(rune(65)))

	// Output:
	// 124 <nil>
	// strconv.Atoi: parsing "12a": invalid syntax
	// 65 A
}

// [end strconv]
