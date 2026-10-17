package sqlbasics

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Priyadharshan0903/gopath-academy/examples/thirdparty/internal/testdb"
)

func setup(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(context.Background(), testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := CreateSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCRUD(t *testing.T) {
	db := setup(t)
	ctx := t.Context()

	id, err := Insert(ctx, db, User{Name: "Ada", Email: "ada@example.com", Nickname: sql.Null[string]{V: "countess", Valid: true}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Insert(ctx, db, User{Name: "Alan", Email: "alan@example.com"}); err != nil {
		t.Fatal(err)
	}

	u, err := Get(ctx, db, id)
	if err != nil || u.Name != "Ada" || !u.Nickname.Valid || u.Nickname.V != "countess" {
		t.Fatalf("Get = %+v, %v", u, err)
	}
	if _, err := Get(ctx, db, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(999) err = %v, want ErrNotFound", err)
	}

	users, err := List(ctx, db)
	if err != nil || len(users) != 2 || users[1].Nickname.Valid {
		t.Fatalf("List = %+v, %v (Alan's nickname should be NULL)", users, err)
	}
}

// [start injectiontest]

func TestInjection(t *testing.T) {
	db := setup(t)
	ctx := t.Context()
	Insert(ctx, db, User{Name: "Ada", Email: "ada@example.com"})
	Insert(ctx, db, User{Name: "Alan", Email: "alan@example.com"})

	evil := `nobody' OR '1'='1`
	unsafe, _ := FindByNameUnsafe(ctx, db, evil)
	safe, _ := FindByName(ctx, db, evil)
	t.Logf("unsafe query matched %d users, safe query matched %d", unsafe, safe)
	if unsafe != 2 || safe != 0 {
		t.Fatalf("unsafe=%d safe=%d", unsafe, safe)
	}
}

// [end injectiontest]

// [start pooltest]

func TestPoolLimit(t *testing.T) {
	db := setup(t)
	db.SetMaxOpenConns(2) // only two queries can run at once

	start := time.Now()
	var wg sync.WaitGroup
	for range 6 {
		wg.Go(func() {
			if _, err := db.ExecContext(t.Context(), `SELECT pg_sleep(0.2)`); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	elapsed := time.Since(start)
	s := db.Stats()

	t.Logf("6 queries of 200ms with 2 connections took %v; waited %d times for %v in total",
		elapsed.Round(10*time.Millisecond), s.WaitCount, s.WaitDuration.Round(10*time.Millisecond))
	if elapsed < 600*time.Millisecond || s.WaitCount < 4 {
		t.Fatalf("expected queuing: elapsed %v, waits %d", elapsed, s.WaitCount)
	}
}

// [end pooltest]
