package kv

import "testing"

// Compile-time checks: the build fails if a type stops satisfying Store.
var (
	_ Store = (*MemStore)(nil)
	_ Store = Prefixed{}
	_ Store = (*Counting)(nil)
)

func TestMemStore(t *testing.T) {
	m := NewMemStore()
	if _, ok := m.Get("a"); ok {
		t.Error("empty store reported key a as present")
	}
	m.Set("a", "1")
	m.Set("a", "2")
	if v, ok := m.Get("a"); !ok || v != "2" {
		t.Errorf("Get(a) = %q, %v; want 2, true", v, ok)
	}
	m.Set("empty", "")
	if v, ok := m.Get("empty"); !ok || v != "" {
		t.Errorf("an empty value must still be reported as present, got %q, %v", v, ok)
	}
}

func TestPrefixed(t *testing.T) {
	mem := NewMemStore()
	users := Prefixed{Prefix: "user:", Inner: mem}
	orders := Prefixed{Prefix: "order:", Inner: mem}

	users.Set("1", "ana")
	orders.Set("1", "book")

	if v, _ := users.Get("1"); v != "ana" {
		t.Errorf("users.Get(1) = %q, want ana", v)
	}
	if v, _ := orders.Get("1"); v != "book" {
		t.Errorf("orders.Get(1) = %q, want book", v)
	}
	if v, ok := mem.Get("user:1"); !ok || v != "ana" {
		t.Errorf("inner store should hold user:1 = ana, got %q, %v", v, ok)
	}
}

func TestCounting(t *testing.T) {
	c := &Counting{Inner: NewMemStore()}
	c.Set("k", "v")
	c.Get("k")
	c.Get("k")
	c.Get("missing")
	if c.Sets != 1 || c.Hits != 2 || c.Misses != 1 {
		t.Errorf("Sets=%d Hits=%d Misses=%d, want 1, 2, 1", c.Sets, c.Hits, c.Misses)
	}
}

func TestDecoratorsCompose(t *testing.T) {
	// Any Store can wrap any other Store.
	counter := &Counting{Inner: NewMemStore()}
	var s Store = Prefixed{Prefix: "x:", Inner: counter}
	s.Set("a", "1")
	if v, ok := s.Get("a"); !ok || v != "1" {
		t.Fatalf("composed Get = %q, %v", v, ok)
	}
	if counter.Sets != 1 || counter.Hits != 1 {
		t.Errorf("inner counter saw Sets=%d Hits=%d, want 1, 1", counter.Sets, counter.Hits)
	}
}
