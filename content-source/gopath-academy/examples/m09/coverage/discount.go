// Package coverage holds the examples for lesson 9.8.
package coverage

// [start code]

// Discount returns the discount percentage for an order.
func Discount(total int, member bool, coupon string) int {
	d := 0
	if total >= 10000 {
		d = 10
	} else if total >= 5000 {
		d = 5
	}
	if member {
		d += 5
	}
	switch coupon {
	case "DIWALI":
		d += 15
	case "WELCOME":
		if d == 0 {
			d = 10
		}
	}
	return min(d, 30)
}

// [end code]
