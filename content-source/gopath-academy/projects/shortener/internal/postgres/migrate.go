package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The SQL files are compiled into the binary, so a deployment is one
// file and the code can never run against migrations from another
// version.
//
//go:embed migrations/*.sql
var migrations embed.FS

// migrationLock is an arbitrary number identifying our advisory lock.
const migrationLock = 727_001

// Migrate applies every migration that has not been applied yet, in
// file name order, and returns the names of those it applied.
//
// All of it runs in one transaction (PostgreSQL can roll back schema
// changes), holding an advisory lock, so several replicas starting at
// the same time cannot apply a migration twice.
func Migrate(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	files, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return nil, err
	}
	slices.Sort(files)

	var applied []string
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, migrationLock); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `SELECT version FROM schema_migrations`)
		if err != nil {
			return err
		}
		done, err := pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			return err
		}

		for _, f := range files {
			version := path.Base(f)
			if slices.Contains(done, version) {
				continue
			}
			sql, err := migrations.ReadFile(f)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, string(sql)); err != nil {
				return fmt.Errorf("migration %s: %w", version, err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
				return err
			}
			applied = append(applied, version)
		}
		return nil
	})
	return applied, err
}
