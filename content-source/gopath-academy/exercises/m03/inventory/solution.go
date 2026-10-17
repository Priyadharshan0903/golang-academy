//go:build solution

// Package inventory practices maps: lookups, the comma-ok form,
// deleting, copying and producing sorted output.
package inventory

import (
	"maps"
	"slices"
)

// Add increases the quantity of item in stock by n.
func Add(stock map[string]int, item string, n int) {
	stock[item] += n
}

// Remove takes n units of item out of stock.
func Remove(stock map[string]int, item string, n int) bool {
	have, ok := stock[item]
	if !ok || have < n {
		return false
	}
	if have == n {
		delete(stock, item)
	} else {
		stock[item] = have - n
	}
	return true
}

// Merge returns a new map with the quantities of a and b added together.
func Merge(a, b map[string]int) map[string]int {
	out := maps.Clone(a)
	if out == nil {
		out = make(map[string]int, len(b))
	}
	for k, v := range b {
		out[k] += v
	}
	return out
}

// LowStock returns items below threshold, sorted alphabetically.
func LowStock(stock map[string]int, threshold int) []string {
	var low []string
	for item, qty := range stock {
		if qty < threshold {
			low = append(low, item)
		}
	}
	slices.Sort(low)
	return low
}
