package txdemo

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/examples/thirdparty/internal/testdb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setup(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testdb.Pool(t)
	if err := CreateSchema(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	return pool
}

func balance(t *testing.T, pool *pgxpool.Pool, id string) int64 {
	t.Helper()
	var b int64
	if err := pool.QueryRow(t.Context(), `SELECT balance FROM accounts WHERE id = $1`, id).Scan(&b); err != nil {
		t.Fatal(err)
	}
	return b
}

// [start losttest]

func TestLostUpdate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		withdraw func(context.Context, *pgxpool.Pool, string, int64, func()) error
		want     int64
	}{
		{"racy", WithdrawRacy, 90},     // one withdrawal is lost
		{"locked", WithdrawLocked, 80}, // both count
	} {
		t.Run(tc.name, func(t *testing.T) {
			pool := setup(t)
			var wg sync.WaitGroup
			for range 2 {
				wg.Go(func() {
					pause := func() { time.Sleep(100 * time.Millisecond) }
					if err := tc.withdraw(t.Context(), pool, "alice", 10, pause); err != nil {
						t.Error(err)
					}
				})
			}
			wg.Wait()
			got := balance(t, pool, "alice")
			t.Logf("two withdrawals of 10 from 100 left %d", got)
			if got != tc.want {
				t.Fatalf("balance %d, want %d", got, tc.want)
			}
		})
	}
}

// [end losttest]

func TestTransfer(t *testing.T) {
	pool := setup(t)
	ctx := t.Context()
	if err := Transfer(ctx, pool, "alice", "bob", 30); err != nil {
		t.Fatal(err)
	}
	err := Transfer(ctx, pool, "alice", "bob", 1000)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("err = %v", err)
	}
	if a, b := balance(t, pool, "alice"), balance(t, pool, "bob"); a != 70 || b != 30 {
		t.Fatalf("alice=%d bob=%d, want 70 30 (the failed transfer changed nothing)", a, b)
	}
}

// [start deadlocktest]

// concurrentTransfers runs 20 transfers of 1, half in each direction,
// and counts the errors by SQLSTATE code.
func concurrentTransfers(t *testing.T, pool *pgxpool.Pool,
	transfer func(context.Context, *pgxpool.Pool, string, string, int64) error) map[string]int {
	var mu sync.Mutex
	codes := map[string]int{}
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			from, to := "alice", "bob"
			if i%2 == 0 {
				from, to = to, from
			}
			err := transfer(t.Context(), pool, from, to, 1)
			var pgErr *pgconn.PgError
			mu.Lock()
			defer mu.Unlock()
			switch {
			case errors.As(err, &pgErr):
				codes[pgErr.Code]++
			case errors.Is(err, ErrInsufficientFunds):
				codes["insufficient funds"]++
			case err != nil:
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if total := balance(t, pool, "alice") + balance(t, pool, "bob"); total != 100 {
		t.Fatalf("total %d, want 100: money was created or destroyed", total)
	}
	return codes
}

func TestDeadlock(t *testing.T) {
	naive := concurrentTransfers(t, setup(t), TransferNaive)
	t.Logf("naive transfers, errors by code: %v (40P01 is deadlock_detected)", naive)

	ordered := concurrentTransfers(t, setup(t), Transfer)
	t.Logf("ordered locking, errors by code: %v", ordered)
	if ordered["40P01"] != 0 {
		t.Fatalf("ordered locking deadlocked %d times", ordered["40P01"])
	}
}

// [end deadlocktest]

// [start rollbacktest]

func TestRollbackOnError(t *testing.T) {
	pool := setup(t)
	err := WithTx(t.Context(), pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(t.Context(), `UPDATE accounts SET balance = 0 WHERE id = 'alice'`); err != nil {
			return err
		}
		return errors.New("something failed after the update")
	})
	if err == nil || balance(t, pool, "alice") != 100 {
		t.Fatalf("err=%v balance=%d: the update should have been rolled back", err, balance(t, pool, "alice"))
	}
}

// [end rollbacktest]

// [start canceltest]

func TestContextCancelsQuery(t *testing.T) {
	pool := setup(t)
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := pool.Exec(ctx, `SELECT pg_sleep(5)`)
	elapsed := time.Since(start)
	t.Logf("a 5 s query with a 100 ms deadline returned after %v: %v", elapsed.Round(10*time.Millisecond), err)
	if err == nil || elapsed > 2*time.Second {
		t.Fatal("the query should have been cancelled")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("errors.Is(err, context.DeadlineExceeded) = false")
	}
}

// [end canceltest]
