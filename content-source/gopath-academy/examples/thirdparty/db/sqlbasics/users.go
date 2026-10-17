// Package sqlbasics holds the database/sql examples for lesson 13.1.
// It uses PostgreSQL through pgx's database/sql driver.
package sqlbasics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver with database/sql
)

// [start open]

// Open returns a *sql.DB, which is a pool of connections, not a single
// connection. It is safe for concurrent use and meant to live for the
// whole program.
func Open(ctx context.Context, url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url) // validates arguments only; does not connect
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)                  // never more than 10 connections at once
	db.SetMaxIdleConns(10)                  // keep up to 10 open between bursts
	db.SetConnMaxLifetime(30 * time.Minute) // recycle connections periodically
	db.SetConnMaxIdleTime(5 * time.Minute)  // close connections idle this long

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil { // the first real connection
		db.Close()
		return nil, fmt.Errorf("connect: %w", err)
	}
	return db, nil
}

// [end open]

// [start crud]

var ErrNotFound = errors.New("not found")

type User struct {
	ID       int64
	Name     string
	Email    string
	Nickname sql.Null[string] // a nullable column (sql.Null[T] is Go 1.22+)
}

func CreateSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE users (
			id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name     text NOT NULL,
			email    text NOT NULL UNIQUE,
			nickname text
		)`)
	return err
}

// Insert adds a user and returns its new ID. Values are passed as
// parameters ($1, $2, $3), never formatted into the SQL string.
func Insert(ctx context.Context, db *sql.DB, u User) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx,
		`INSERT INTO users (name, email, nickname) VALUES ($1, $2, $3) RETURNING id`,
		u.Name, u.Email, u.Nickname,
	).Scan(&id)
	return id, err
}

// Get returns one user, or ErrNotFound.
func Get(ctx context.Context, db *sql.DB, id int64) (User, error) {
	var u User
	err := db.QueryRowContext(ctx,
		`SELECT id, name, email, nickname FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Nickname)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, fmt.Errorf("user %d: %w", id, ErrNotFound)
	}
	return u, err
}

// List returns all users ordered by name.
func List(ctx context.Context, db *sql.DB) ([]User, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name, email, nickname FROM users ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // returns the connection to the pool

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Nickname); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err() // an error that ended the loop early
}

// [end crud]

// [start injection]

// FindByNameUnsafe builds SQL with string formatting. Never do this.
func FindByNameUnsafe(ctx context.Context, db *sql.DB, name string) (int, error) {
	var n int
	err := db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT count(*) FROM users WHERE name = '%s'`, name)).Scan(&n)
	return n, err
}

// FindByName passes name as a parameter: the database never parses it
// as SQL, whatever it contains.
func FindByName(ctx context.Context, db *sql.DB, name string) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE name = $1`, name).Scan(&n)
	return n, err
}

// [end injection]
