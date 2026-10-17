//go:build !solution

// Package tempconv converts between Celsius and Fahrenheit.
//
// The tests live in package tempconv_test (an external test package),
// so they can only call names you export. That is the point of this
// exercise: decide what is public API and what is an internal detail.
package tempconv

// AbsoluteZeroC is the lowest possible temperature in Celsius.
const AbsoluteZeroC = -273.15

// CToF converts Celsius to Fahrenheit: F = C*9/5 + 32.
func CToF(c float64) float64 {
	// TODO: implement me.
	return 0
}

// FToC converts Fahrenheit to Celsius: C = (F - 32) * 5/9.
func FToC(f float64) float64 {
	// TODO: implement me.
	return 0
}

// IsPhysical reports whether c is at or above absolute zero.
func IsPhysical(c float64) bool {
	// TODO: implement me.
	return false
}
