//go:build !solution

// Package inventory practices maps: lookups, the comma-ok form,
// deleting, copying and producing sorted output.
package inventory

// Add increases the quantity of item in stock by n.
// stock must not be nil.
func Add(stock map[string]int, item string, n int) {
	// TODO: implement me.
}

// Remove takes n units of item out of stock. It returns false, and
// changes nothing, if the item is missing or there are fewer than n.
// When the quantity reaches 0, the item is deleted from the map.
func Remove(stock map[string]int, item string, n int) bool {
	// TODO: implement me.
	return false
}

// Merge returns a NEW map with the quantities of a and b added
// together. It must not modify a or b. Either may be nil.
func Merge(a, b map[string]int) map[string]int {
	// TODO: implement me.
	return nil
}

// LowStock returns the names of items whose quantity is below
// threshold, sorted alphabetically.
func LowStock(stock map[string]int, threshold int) []string {
	// TODO: implement me.
	return nil
}
