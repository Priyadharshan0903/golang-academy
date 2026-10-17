// Package color is the input of the enumgen example (lesson 14.3).
// Run `go generate ./examples/m14/...` after changing the constants.
package color

//go:generate go run ../enumgen -type=Color

// Color is a traffic light color.
type Color int

const (
	ColorRed Color = iota + 1 // start at 1, so the zero value is not a valid color
	ColorAmber
	ColorGreen
)
