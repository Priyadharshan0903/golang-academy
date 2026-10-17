// Package pgtest gives each test its own empty PostgreSQL schema.
//
// Tests that need a database call pgtest.URL or pgtest.Pool. They are
// skipped unless TEST_DATABASE_URL is set, for example:
//
//	docker run -d --name pg -e POSTGRES_PASSWORD=pg -p 5432:5432 postgres:16
//	export TEST_DATABASE_URL='postgres://postgres:pg@localhost:5432/postgres?sslmode=disable'
package pgtest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// URL creates a fresh schema for t and returns a connection URL whose
// search_path points at it. The schema is dropped when t ends.
func URL(t testing.TB) string {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL not set; see projects/shortener/README.md")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, base)
	if err != nil {
		t.Fatalf("connect to TEST_DATABASE_URL: %v", err)
	}
	defer conn.Close(ctx)

	// The process ID keeps parallel test binaries (one per package)
	// apart; PostgreSQL identifiers are limited to 63 bytes.
	name := strings.ToLower(strings.NewReplacer("/", "_", "#", "_", " ", "_").Replace(t.Name()))
	if len(name) > 40 {
		name = name[:40]
	}
	schema := pgx.Identifier{fmt.Sprintf("test_%s_%d", name, os.Getpid())}.Sanitize()
	if _, err := conn.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE; CREATE SCHEMA %s", schema, schema)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), base)
		if err == nil {
			c.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
			c.Close(context.Background())
		}
	})

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", strings.Trim(schema, `"`))
	u.RawQuery = q.Encode()
	return u.String()
}

// Pool returns a pgxpool connected to a fresh schema for t.
func Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}
