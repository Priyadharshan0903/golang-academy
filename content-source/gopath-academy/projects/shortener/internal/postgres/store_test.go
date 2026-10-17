package postgres

import (
	"context"
	"sync"
	"testing"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link/linktest"
	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/pgtest"
)

func TestStoreContract(t *testing.T) {
	linktest.Run(t, func(t *testing.T) link.Store {
		pool := pgtest.Pool(t)
		if _, err := Migrate(context.Background(), pool); err != nil {
			t.Fatal(err)
		}
		return NewStore(pool)
	})
}

func TestMigrateIsIdempotent(t *testing.T) {
	pool := pgtest.Pool(t)
	first, err := Migrate(t.Context(), pool)
	if err != nil || len(first) == 0 {
		t.Fatalf("first run applied %v, err %v", first, err)
	}
	second, err := Migrate(t.Context(), pool)
	if err != nil || len(second) != 0 {
		t.Fatalf("second run applied %v, err %v; want nothing", second, err)
	}
	t.Logf("applied %v, then nothing", first)
}

// Several replicas starting at once must not apply a migration twice.
func TestMigrateConcurrently(t *testing.T) {
	pool := pgtest.Pool(t)
	var mu sync.Mutex
	total := 0
	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			applied, err := Migrate(t.Context(), pool)
			if err != nil {
				t.Error(err)
			}
			mu.Lock()
			total += len(applied)
			mu.Unlock()
		})
	}
	wg.Wait()
	files, _ := migrations.ReadDir("migrations")
	if total != len(files) {
		t.Fatalf("migrations applied %d times in total, want %d", total, len(files))
	}
}
