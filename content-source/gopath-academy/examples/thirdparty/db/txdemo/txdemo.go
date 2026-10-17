// Package txdemo holds the transaction examples for lesson 13.4.
package txdemo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE accounts (
			id      text PRIMARY KEY,
			balance bigint NOT NULL CHECK (balance >= 0)
		);
		INSERT INTO accounts VALUES ('alice', 100), ('bob', 0);`)
	return err
}

// [start withtx]

// WithTx runs fn in a transaction. It commits if fn returns nil and
// rolls back otherwise, including when fn panics.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) (err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	// Rollback after a successful Commit is a no-op, so this is safe to
	// defer unconditionally.
	defer tx.Rollback(context.WithoutCancel(ctx))

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// [end withtx]

// [start lost]

// WithdrawRacy reads the balance, then writes back a new value. Two of
// these running at the same time can both read 100 and both write 90:
// a lost update. pause widens the gap between read and write so the
// test can make the race happen reliably.
func WithdrawRacy(ctx context.Context, pool *pgxpool.Pool, id string, amount int64, pause func()) error {
	return WithTx(ctx, pool, func(tx pgx.Tx) error {
		var balance int64
		if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, id).Scan(&balance); err != nil {
			return err
		}
		pause()
		_, err := tx.Exec(ctx, `UPDATE accounts SET balance = $1 WHERE id = $2`, balance-amount, id)
		return err
	})
}

// WithdrawLocked does the same, but SELECT ... FOR UPDATE locks the
// row: a second transaction waits at its SELECT until the first commits,
// then reads the new balance.
func WithdrawLocked(ctx context.Context, pool *pgxpool.Pool, id string, amount int64, pause func()) error {
	return WithTx(ctx, pool, func(tx pgx.Tx) error {
		var balance int64
		if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1 FOR UPDATE`, id).Scan(&balance); err != nil {
			return err
		}
		pause()
		_, err := tx.Exec(ctx, `UPDATE accounts SET balance = $1 WHERE id = $2`, balance-amount, id)
		return err
	})
}

// [end lost]

// [start transfer]

var ErrInsufficientFunds = errors.New("insufficient funds")

// TransferNaive moves amount between accounts atomically: both updates
// happen, or neither does. Each UPDATE computes the new balance inside
// the database, so there is no read-then-write race. But it can
// deadlock: see TestDeadlock.
func TransferNaive(ctx context.Context, pool *pgxpool.Pool, from, to string, amount int64) error {
	return WithTx(ctx, pool, func(tx pgx.Tx) error {
		return move(ctx, tx, from, to, amount)
	})
}

// Transfer first locks both rows in a fixed order (by id), so two
// transfers in opposite directions can never wait for each other.
func Transfer(ctx context.Context, pool *pgxpool.Pool, from, to string, amount int64) error {
	return WithTx(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`SELECT id FROM accounts WHERE id = ANY($1) ORDER BY id FOR UPDATE`,
			[]string{from, to}); err != nil {
			return err
		}
		return move(ctx, tx, from, to, amount)
	})
}

func move(ctx context.Context, tx pgx.Tx, from, to string, amount int64) error {
	tag, err := tx.Exec(ctx,
		`UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1`, amount, from)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("transfer %d from %s: %w", amount, from, ErrInsufficientFunds)
	}
	_, err = tx.Exec(ctx, `UPDATE accounts SET balance = balance + $1 WHERE id = $2`, amount, to)
	return err
}

// [end transfer]
