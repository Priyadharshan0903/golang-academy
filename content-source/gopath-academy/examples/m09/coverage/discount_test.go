package coverage

import (
	"os"
	"testing"
)

// [start tests]

func TestDiscount(t *testing.T) {
	tests := []struct {
		total  int
		member bool
		coupon string
		want   int
	}{
		{100, false, "", 0},
		{6000, false, "", 5},
		{12000, true, "", 15},
		{12000, true, "DIWALI", 30}, // capped
		{100, false, "WELCOME", 10},
	}
	for _, tt := range tests {
		if got := Discount(tt.total, tt.member, tt.coupon); got != tt.want {
			t.Errorf("Discount(%d, %v, %q) = %d, want %d", tt.total, tt.member, tt.coupon, got, tt.want)
		}
	}
}

// [end tests]

// [start short]

// Slow or external tests opt out under -short, so `go test -short`
// stays fast in the edit-test loop.
func TestAgainstPricingService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	if os.Getenv("PRICING_URL") == "" {
		t.Skip("PRICING_URL not set")
	}
	// ... call the real service here ...
}

// [end short]
