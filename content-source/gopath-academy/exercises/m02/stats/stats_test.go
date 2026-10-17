package stats

import (
	"math"
	"testing"
)

func TestSum(t *testing.T) {
	if got := Sum(); got != 0 {
		t.Errorf("Sum() = %d, want 0", got)
	}
	if got := Sum(1, 2, 3, 4); got != 10 {
		t.Errorf("Sum(1, 2, 3, 4) = %d, want 10", got)
	}
	nums := []int{-5, 5, 100}
	if got := Sum(nums...); got != 100 {
		t.Errorf("Sum(nums...) = %d, want 100", got)
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		name             string
		in               []int
		wantMin, wantMax int
		wantOK           bool
	}{
		{"empty", nil, 0, 0, false},
		{"one", []int{7}, 7, 7, true},
		{"mixed", []int{3, -2, 9, 0}, -2, 9, true},
		{"all equal", []int{4, 4, 4}, 4, 4, true},
		{"negatives", []int{-1, -8, -3}, -8, -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lo, hi, ok := MinMax(tt.in...)
			if lo != tt.wantMin || hi != tt.wantMax || ok != tt.wantOK {
				t.Errorf("MinMax(%v) = (%d, %d, %v), want (%d, %d, %v)",
					tt.in, lo, hi, ok, tt.wantMin, tt.wantMax, tt.wantOK)
			}
		})
	}
}

func TestRunningAverage(t *testing.T) {
	avg := RunningAverage()
	for i, tc := range []struct{ in, want float64 }{
		{10, 10}, {20, 15}, {30, 20}, {-60, 0},
	} {
		if got := avg(tc.in); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("call %d: avg(%v) = %v, want %v", i+1, tc.in, got, tc.want)
		}
	}
}

func TestRunningAverageIndependent(t *testing.T) {
	a, b := RunningAverage(), RunningAverage()
	a(100)
	if got := b(2); got != 2 {
		t.Errorf("second average was affected by the first: got %v, want 2", got)
	}
}
