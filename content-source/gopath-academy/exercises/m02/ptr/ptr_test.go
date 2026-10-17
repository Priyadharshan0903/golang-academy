package ptr

import "testing"

func TestSwap(t *testing.T) {
	x, y := 1, 2
	Swap(&x, &y)
	if x != 2 || y != 1 {
		t.Errorf("after Swap: x=%d y=%d, want x=2 y=1", x, y)
	}

	same := 5
	Swap(&same, &same)
	if same != 5 {
		t.Errorf("swapping a value with itself changed it to %d", same)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct{ in, lo, hi, want int }{
		{5, 0, 10, 5},
		{-3, 0, 10, 0},
		{42, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}
	for _, tt := range tests {
		v := tt.in
		Clamp(&v, tt.lo, tt.hi)
		if v != tt.want {
			t.Errorf("Clamp(%d, %d, %d) gave %d, want %d", tt.in, tt.lo, tt.hi, v, tt.want)
		}
	}
}

func TestClampNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Clamp(nil, ...) panicked: %v", r)
		}
	}()
	Clamp(nil, 0, 10)
}

func TestValueOr(t *testing.T) {
	n := 7
	if got := ValueOr(&n, 1); got != 7 {
		t.Errorf("ValueOr(&7, 1) = %d, want 7", got)
	}
	if got := ValueOr(nil, 1); got != 1 {
		t.Errorf("ValueOr(nil, 1) = %d, want 1", got)
	}
}

func TestIntPtr(t *testing.T) {
	p := IntPtr(3)
	if p == nil {
		t.Fatal("IntPtr returned nil")
	}
	if *p != 3 {
		t.Errorf("*IntPtr(3) = %d, want 3", *p)
	}
	q := IntPtr(3)
	if p == q {
		t.Error("IntPtr must return a new pointer each call")
	}
	*p = 9
	if *q != 3 {
		t.Error("changing one pointer affected another")
	}
}
