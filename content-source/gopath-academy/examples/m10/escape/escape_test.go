package escape

import "testing"

var n1000, n7 = 1000, 7 // variables, so the compiler cannot use static data

// [start test]

func TestAllocations(t *testing.T) {
	cases := []struct {
		name string
		f    func()
		want float64
	}{
		{"value return", func() { _ = NewPointValue(1) }, 0},
		{"pointer return", func() { _ = NewPointPtr(1) }, 1},
		{"escaping interface holding 1000", func() { Remember(n1000) }, 1},
		{"escaping interface holding 7", func() { Remember(n7) }, 0}, // the runtime caches 0..255
		{"non-escaping interface", func() { _ = IsInt(n1000) }, 0},   // boxed on the stack
		{"constant in an interface", func() { Remember(1000) }, 0},   // static read-only data
		{"fixed-size local array", func() { _ = SumFixed() }, 0},
	}
	for _, c := range cases {
		if got := testing.AllocsPerRun(100, c.f); got != c.want {
			t.Errorf("%s: %v allocations, want %v", c.name, got, c.want)
		}
	}
}

// [end test]
