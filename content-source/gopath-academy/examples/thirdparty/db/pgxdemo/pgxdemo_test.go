package pgxdemo

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/examples/thirdparty/internal/testdb"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setup(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := NewPool(context.Background(), testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := CreateSchema(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestByTagAndErrors(t *testing.T) {
	pool := setup(t)
	ctx := t.Context()
	for _, p := range []Product{
		{SKU: "GO-MUG", Name: "Gopher mug", PriceCents: 1200, Tags: []string{"kitchen", "gopher"}},
		{SKU: "GO-TEE", Name: "Gopher t-shirt", PriceCents: 2000, Tags: []string{"clothes", "gopher"}},
		{SKU: "K8S-CAP", Name: "Helm cap", PriceCents: 1500, Tags: []string{"clothes"}},
	} {
		if _, err := Insert(ctx, pool, p); err != nil {
			t.Fatal(err)
		}
	}

	got, err := ByTag(ctx, pool, "gopher")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].SKU != "GO-MUG" || got[1].Tags[0] != "clothes" || got[0].CreatedAt.IsZero() {
		t.Fatalf("ByTag = %+v", got)
	}

	_, err = Insert(ctx, pool, Product{SKU: "GO-MUG", Name: "dup", PriceCents: 1})
	t.Logf("duplicate: %v", err)
	if !errors.Is(err, ErrDuplicateSKU) {
		t.Fatalf("err = %v, want ErrDuplicateSKU", err)
	}
	_, err = Insert(ctx, pool, Product{SKU: "NEG", Name: "negative", PriceCents: -1})
	t.Logf("check violation: %v", err)
	if err == nil || errors.Is(err, ErrDuplicateSKU) {
		t.Fatalf("err = %v, want a check constraint error", err)
	}
}

// [start bulktest]

func TestBulkInsert(t *testing.T) {
	const n = 5000
	for _, tc := range []struct {
		name   string
		insert func(context.Context, *pgxpool.Pool, []Product) error
	}{
		{"one by one", InsertOneByOne},
		{"batch", InsertBatch},
		{"copy", InsertCopy},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := setup(t)
			ps := make([]Product, n)
			for i := range ps {
				ps[i] = Product{SKU: fmt.Sprintf("SKU-%05d", i), Name: "item", PriceCents: int64(i)}
			}
			start := time.Now()
			if err := tc.insert(t.Context(), pool, ps); err != nil {
				t.Fatal(err)
			}
			elapsed := time.Since(start)

			var count int
			pool.QueryRow(t.Context(), `SELECT count(*) FROM products`).Scan(&count)
			t.Logf("%-10s %d rows in %v", tc.name, count, elapsed.Round(time.Millisecond))
			if count != n {
				t.Fatalf("count = %d", count)
			}
		})
	}
}

// [end bulktest]
