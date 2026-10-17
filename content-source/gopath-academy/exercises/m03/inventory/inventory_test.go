package inventory

import (
	"maps"
	"slices"
	"testing"
)

func TestAdd(t *testing.T) {
	stock := map[string]int{"apple": 2}
	Add(stock, "apple", 3)
	Add(stock, "pear", 1)
	want := map[string]int{"apple": 5, "pear": 1}
	if !maps.Equal(stock, want) {
		t.Errorf("stock = %v, want %v", stock, want)
	}
}

func TestRemove(t *testing.T) {
	stock := map[string]int{"apple": 5, "pear": 1}

	if !Remove(stock, "apple", 2) || stock["apple"] != 3 {
		t.Errorf("Remove 2 apples: stock = %v", stock)
	}
	if Remove(stock, "apple", 10) {
		t.Error("removing more than available should fail")
	}
	if stock["apple"] != 3 {
		t.Errorf("failed Remove changed stock: %v", stock)
	}
	if Remove(stock, "kiwi", 1) {
		t.Error("removing a missing item should fail")
	}
	if _, exists := stock["kiwi"]; exists {
		t.Error("failed Remove must not add the key")
	}
	if !Remove(stock, "pear", 1) {
		t.Error("removing the last pear should succeed")
	}
	if _, exists := stock["pear"]; exists {
		t.Error("an item with quantity 0 should be deleted")
	}
}

func TestMerge(t *testing.T) {
	a := map[string]int{"apple": 1, "pear": 2}
	b := map[string]int{"apple": 10, "kiwi": 3}
	got := Merge(a, b)
	want := map[string]int{"apple": 11, "pear": 2, "kiwi": 3}
	if !maps.Equal(got, want) {
		t.Errorf("Merge = %v, want %v", got, want)
	}
	if a["apple"] != 1 || len(a) != 2 || b["apple"] != 10 || len(b) != 2 {
		t.Errorf("Merge modified its inputs: a=%v b=%v", a, b)
	}
	if got := Merge(nil, b); !maps.Equal(got, b) {
		t.Errorf("Merge(nil, b) = %v", got)
	}
	got = Merge(a, nil)
	got["apple"] = 999
	if a["apple"] != 1 {
		t.Error("Merge(a, nil) returned a itself instead of a copy")
	}
	if Merge(nil, nil) == nil {
		t.Error("Merge(nil, nil) should return an empty, usable map")
	}
}

func TestLowStock(t *testing.T) {
	stock := map[string]int{"pear": 1, "apple": 0, "kiwi": 9, "fig": 2}
	got := LowStock(stock, 3)
	if !slices.Equal(got, []string{"apple", "fig", "pear"}) {
		t.Errorf("LowStock = %v, want [apple fig pear]", got)
	}
	if got := LowStock(stock, 0); len(got) != 0 {
		t.Errorf("LowStock(0) = %v, want none", got)
	}
}
