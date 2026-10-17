// Package linktest is a conformance test suite for link.Store
// implementations. Every store (memory, Postgres, cached) runs the same
// tests, so they all behave the way link.Service expects.
package linktest

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
)

// Run runs the suite. newStore must return an empty store.
func Run(t *testing.T, newStore func(t *testing.T) link.Store) {
	t.Run("CreateThenGet", func(t *testing.T) {
		s := newStore(t)
		ctx := context.Background()
		want := link.Link{Code: "abc", URL: "https://go.dev", CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 6000, time.UTC)}
		if err := s.Create(ctx, want); err != nil {
			t.Fatal(err)
		}
		got, err := s.Get(ctx, "abc")
		if err != nil {
			t.Fatal(err)
		}
		// Stores may keep only microseconds (PostgreSQL does) and may
		// return another time zone; Equal compares instants.
		if got.Code != want.Code || got.URL != want.URL || !got.CreatedAt.Equal(want.CreatedAt) {
			t.Fatalf("Get = %+v, want %+v", got, want)
		}
		if got.CreatedAt.Location() != time.UTC {
			t.Errorf("CreatedAt should be in UTC, got %v", got.CreatedAt.Location())
		}
	})

	t.Run("GetMissing", func(t *testing.T) {
		_, err := newStore(t).Get(context.Background(), "nope")
		if !errors.Is(err, link.ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("CreateDuplicate", func(t *testing.T) {
		s := newStore(t)
		ctx := context.Background()
		l := link.Link{Code: "dup", URL: "https://a.example", CreatedAt: time.Now().UTC()}
		if err := s.Create(ctx, l); err != nil {
			t.Fatal(err)
		}
		l.URL = "https://b.example"
		if err := s.Create(ctx, l); !errors.Is(err, link.ErrCodeTaken) {
			t.Fatalf("err = %v, want ErrCodeTaken", err)
		}
		got, _ := s.Get(ctx, "dup")
		if got.URL != "https://a.example" {
			t.Fatalf("the duplicate overwrote the original: %+v", got)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		s := newStore(t)
		ctx := context.Background()
		l := link.Link{Code: "gone", URL: "https://go.dev", CreatedAt: time.Now().UTC()}
		if err := s.Create(ctx, l); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Get(ctx, "gone"); err != nil { // with a cache, this fills it
			t.Fatal(err)
		}
		if err := s.Delete(ctx, "gone"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Get(ctx, "gone"); !errors.Is(err, link.ErrNotFound) {
			t.Fatalf("Get after Delete: err = %v, want ErrNotFound", err)
		}
		if err := s.Delete(ctx, "gone"); !errors.Is(err, link.ErrNotFound) {
			t.Fatalf("second Delete: err = %v, want ErrNotFound", err)
		}
		if err := s.Create(ctx, l); err != nil {
			t.Fatalf("the code should be free again after Delete: %v", err)
		}
	})

	t.Run("ConcurrentCreateSameCode", func(t *testing.T) {
		s := newStore(t)
		var ok, taken atomic.Int64
		var wg sync.WaitGroup
		for range 20 {
			wg.Go(func() {
				err := s.Create(context.Background(), link.Link{Code: "race", URL: "https://go.dev", CreatedAt: time.Now().UTC()})
				switch {
				case err == nil:
					ok.Add(1)
				case errors.Is(err, link.ErrCodeTaken):
					taken.Add(1)
				default:
					t.Error(err)
				}
			})
		}
		wg.Wait()
		if ok.Load() != 1 || taken.Load() != 19 {
			t.Fatalf("%d succeeded and %d got ErrCodeTaken, want 1 and 19", ok.Load(), taken.Load())
		}
	})
}
