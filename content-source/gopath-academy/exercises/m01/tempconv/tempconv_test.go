package tempconv_test

import (
	"math"
	"testing"

	"github.com/Priyadharshan0903/gopath-academy/exercises/m01/tempconv"
)

// Floating point results are rarely exact, so compare with a tolerance.
func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCToF(t *testing.T) {
	cases := []struct{ c, f float64 }{
		{0, 32}, {100, 212}, {-40, -40}, {37, 98.6},
	}
	for _, tc := range cases {
		if got := tempconv.CToF(tc.c); !near(got, tc.f) {
			t.Errorf("CToF(%v) = %v, want %v", tc.c, got, tc.f)
		}
	}
}

func TestFToC(t *testing.T) {
	cases := []struct{ f, c float64 }{
		{32, 0}, {212, 100}, {-40, -40}, {98.6, 37},
	}
	for _, tc := range cases {
		if got := tempconv.FToC(tc.f); !near(got, tc.c) {
			t.Errorf("FToC(%v) = %v, want %v", tc.f, got, tc.c)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	for _, c := range []float64{-273.15, -10, 0, 21.5, 1000} {
		if got := tempconv.FToC(tempconv.CToF(c)); !near(got, c) {
			t.Errorf("FToC(CToF(%v)) = %v", c, got)
		}
	}
}

func TestIsPhysical(t *testing.T) {
	if !tempconv.IsPhysical(tempconv.AbsoluteZeroC) {
		t.Error("absolute zero itself should be physical")
	}
	if !tempconv.IsPhysical(25) {
		t.Error("25C should be physical")
	}
	if tempconv.IsPhysical(-300) {
		t.Error("-300C is below absolute zero")
	}
}
