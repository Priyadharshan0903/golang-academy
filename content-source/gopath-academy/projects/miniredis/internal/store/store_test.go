package store

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func newTestStore() (*Store, *time.Time) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := New()
	s.now = func() time.Time { return now }
	return s, &now
}

func TestSetGetDel(t *testing.T) {
	s, _ := newTestStore()
	s.Set("a", "1", 0)
	if v, ok := s.Get("a"); !ok || v != "1" {
		t.Errorf("Get(a) = %q, %v", v, ok)
	}
	if n := s.Del("a", "missing"); n != 1 {
		t.Errorf("Del returned %d, want 1", n)
	}
	if _, ok := s.Get("a"); ok {
		t.Error("a should be gone")
	}
}

func TestExpiry(t *testing.T) {
	s, now := newTestStore()
	s.Set("session", "x", 10*time.Second)
	if got := s.TTL("session"); got != 10*time.Second {
		t.Errorf("TTL = %v", got)
	}
	*now = now.Add(9 * time.Second)
	if _, ok := s.Get("session"); !ok {
		t.Error("key expired too early")
	}
	*now = now.Add(time.Second)
	if _, ok := s.Get("session"); ok {
		t.Error("key should have expired at exactly its deadline")
	}
	if s.TTL("session") != -2 || s.Exists("session") != 0 {
		t.Error("expired key should look missing")
	}
	if s.Len() != 1 {
		t.Error("expired keys are removed lazily, not on read")
	}
	if n := s.RemoveExpired(); n != 1 || s.Len() != 0 {
		t.Errorf("RemoveExpired removed %d", n)
	}
}

func TestTTLWithoutExpiry(t *testing.T) {
	s, _ := newTestStore()
	s.Set("k", "v", 0)
	if s.TTL("k") != -1 || s.TTL("nope") != -2 {
		t.Error("TTL codes wrong")
	}
	if !s.Expire("k", time.Minute) || s.Expire("nope", time.Minute) {
		t.Error("Expire result wrong")
	}
}

func TestIncr(t *testing.T) {
	s, _ := newTestStore()
	for want := int64(1); want <= 3; want++ {
		if got, err := s.Incr("hits"); err != nil || got != want {
			t.Fatalf("Incr = %d, %v; want %d", got, err, want)
		}
	}
	s.Set("name", "ana", 0)
	if _, err := s.Incr("name"); !errors.Is(err, ErrNotInteger) {
		t.Errorf("Incr on text: %v", err)
	}
}

func TestConcurrentIncr(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				s.Incr("counter")
			}
		}()
	}
	wg.Wait()
	if v, _ := s.Get("counter"); v != "5000" {
		t.Errorf("counter = %s, want 5000 (lost updates)", v)
	}
}
