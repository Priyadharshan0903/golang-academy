package variables

import "fmt"

// [start declare]

func Example_declare() {
	var name string = "Ana" // explicit type and value
	var age = 31            // type inferred from the value (int)
	var city string         // no value: gets the zero value ""
	country := "India"      // short declaration, only inside functions

	fmt.Println(name, age, country)
	fmt.Printf("city=%q\n", city)

	// Output:
	// Ana 31 India
	// city=""
}

// [end declare]

// [start zero]

func Example_zeroValues() {
	var (
		i  int
		f  float64
		b  bool
		s  string
		p  *int
		e  error
		m  map[string]int
		sl []int
	)
	fmt.Println(i, f, b, p, e, m, sl)
	fmt.Printf("%q\n", s)
	fmt.Println(m == nil, sl == nil, len(sl))

	// Output:
	// 0 0 false <nil> <nil> map[] []
	// ""
	// true true 0
}

// [end zero]

// [start multi]

func Example_multipleAssignment() {
	x, y := 1, 2
	x, y = y, x // swap without a temporary variable
	fmt.Println(x, y)

	// := needs at least one NEW variable on the left.
	// Existing ones in the same scope are simply assigned.
	a, err := 10, error(nil)
	b, err := 20, fmt.Errorf("boom") // b is new, err is reused
	fmt.Println(a, b, err)

	// Output:
	// 2 1
	// 10 20 boom
}

// [end multi]
