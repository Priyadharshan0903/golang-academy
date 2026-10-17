// Package postgres is a link.Store backed by PostgreSQL, using pgx.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Priyadharshan0903/gopath-academy/projects/shortener/internal/link"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store implements link.Store.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore returns a Store using pool. Run Migrate first.
func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Create inserts l. The primary key on code makes the database reject
// duplicates atomically; that error becomes link.ErrCodeTaken.
func (s *Store) Create(ctx context.Context, l link.Link) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO links (code, url, created_at) VALUES ($1, $2, $3)`,
		l.Code, l.URL, l.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return link.ErrCodeTaken
	}
	if err != nil {
		return fmt.Errorf("insert link: %w", err)
	}
	return nil
}

// Get returns the link for code, or link.ErrNotFound.
func (s *Store) Get(ctx context.Context, code string) (link.Link, error) {
	var l link.Link
	err := s.pool.QueryRow(ctx,
		`SELECT code, url, created_at FROM links WHERE code = $1`, code,
	).Scan(&l.Code, &l.URL, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return link.Link{}, link.ErrNotFound
	}
	if err != nil {
		return link.Link{}, fmt.Errorf("select link: %w", err)
	}
	l.CreatedAt = l.CreatedAt.UTC() // pgx returns timestamptz in the local time zone
	return l, nil
}

// Delete removes the link for code, or returns link.ErrNotFound.
func (s *Store) Delete(ctx context.Context, code string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM links WHERE code = $1`, code)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return link.ErrNotFound
	}
	return nil
}
