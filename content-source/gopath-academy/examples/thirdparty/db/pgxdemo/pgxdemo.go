// Package pgxdemo holds the pgx examples for lesson 13.2.
package pgxdemo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// [start pool]

// NewPool connects with pgx's own pool. Settings can also be given in
// the URL, for example ?pool_max_conns=10.
func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// [end pool]

// [start rows]

// Product maps to a row. The db tags name the columns.
type Product struct {
	ID         int64     `db:"id"`
	SKU        string    `db:"sku"`
	Name       string    `db:"name"`
	PriceCents int64     `db:"price_cents"`
	Tags       []string  `db:"tags"` // a PostgreSQL text[] column
	CreatedAt  time.Time `db:"created_at"`
}

func CreateSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE products (
			id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			sku         text NOT NULL UNIQUE,
			name        text NOT NULL,
			price_cents bigint NOT NULL CHECK (price_cents >= 0),
			tags        text[] NOT NULL DEFAULT '{}',
			created_at  timestamptz NOT NULL DEFAULT now()
		)`)
	return err
}

// ByTag returns the products with a tag, cheapest first. pgx scans
// rows straight into structs, and converts []string to text[].
func ByTag(ctx context.Context, pool *pgxpool.Pool, tag string) ([]Product, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, sku, name, price_cents, tags, created_at
		   FROM products WHERE $1 = ANY(tags) ORDER BY price_cents`, tag)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[Product]) // closes rows
}

// [end rows]

// [start errors]

var ErrDuplicateSKU = errors.New("duplicate SKU")

// Insert adds a product. A unique violation becomes ErrDuplicateSKU;
// other database errors are returned wrapped.
func Insert(ctx context.Context, pool *pgxpool.Pool, p Product) (int64, error) {
	tags := p.Tags
	if tags == nil {
		tags = []string{} // a nil slice is sent as NULL, not as an empty array
	}
	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO products (sku, name, price_cents, tags) VALUES ($1, $2, $3, $4) RETURNING id`,
		p.SKU, p.Name, p.PriceCents, tags).Scan(&id)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return 0, fmt.Errorf("insert %q: %w", p.SKU, ErrDuplicateSKU)
	}
	if err != nil {
		return 0, fmt.Errorf("insert %q: %w", p.SKU, err)
	}
	return id, nil
}

// [end errors]

// [start bulk]

// InsertOneByOne sends one INSERT per row: one network round trip each.
func InsertOneByOne(ctx context.Context, pool *pgxpool.Pool, ps []Product) error {
	for _, p := range ps {
		if _, err := pool.Exec(ctx,
			`INSERT INTO products (sku, name, price_cents) VALUES ($1, $2, $3)`,
			p.SKU, p.Name, p.PriceCents); err != nil {
			return err
		}
	}
	return nil
}

// InsertBatch queues all INSERTs and sends them in one round trip.
func InsertBatch(ctx context.Context, pool *pgxpool.Pool, ps []Product) error {
	var b pgx.Batch
	for _, p := range ps {
		b.Queue(`INSERT INTO products (sku, name, price_cents) VALUES ($1, $2, $3)`,
			p.SKU, p.Name, p.PriceCents)
	}
	return pool.SendBatch(ctx, &b).Close() // Close reports the first error
}

// InsertCopy uses PostgreSQL's COPY protocol, built for bulk loading.
func InsertCopy(ctx context.Context, pool *pgxpool.Pool, ps []Product) error {
	_, err := pool.CopyFrom(ctx,
		pgx.Identifier{"products"},
		[]string{"sku", "name", "price_cents"},
		pgx.CopyFromSlice(len(ps), func(i int) ([]any, error) {
			return []any{ps[i].SKU, ps[i].Name, ps[i].PriceCents}, nil
		}))
	return err
}

// [end bulk]
