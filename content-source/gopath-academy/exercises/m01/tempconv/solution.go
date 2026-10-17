//go:build solution

// Package tempconv converts between Celsius and Fahrenheit.
package tempconv

// AbsoluteZeroC is the lowest possible temperature in Celsius.
const AbsoluteZeroC = -273.15

// CToF converts Celsius to Fahrenheit.
func CToF(c float64) float64 { return c*9/5 + 32 }

// FToC converts Fahrenheit to Celsius.
func FToC(f float64) float64 { return (f - 32) * 5 / 9 }

// IsPhysical reports whether c is at or above absolute zero.
func IsPhysical(c float64) bool { return c >= AbsoluteZeroC }
