package repository

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
	"sync"
	"sync/atomic"
	"testing"
)

// Set TEST_DATABASE_URL to a disposable PostgreSQL database; no production tables are touched.
func TestPasswordResetLifecycle(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TEMP TABLE users (id uuid PRIMARY KEY, email text, password text, reset_token text,
        reset_token_expiry timestamptz, deleted_at timestamptz, updated_at timestamptz)`)
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	if _, err := db.Exec(`INSERT INTO users (id,email,password) VALUES ($1,'test@example.invalid','old')`, id); err != nil {
		t.Fatal(err)
	}
	repo := &UserRepository{DB: db}
	issued, err := repo.IssuePasswordReset(id, "first")
	if err != nil || !issued {
		t.Fatalf("issue: %v", err)
	}
	if issued, err := repo.IssuePasswordReset(id, "second"); err != nil || issued {
		t.Fatalf("cooldown: %v", err)
	}
	if _, err := repo.ConsumePasswordReset("second", "new"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unissued token accepted: %v", err)
	}
	var successful atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := repo.ConsumePasswordReset("first", "new"); err == nil {
				successful.Add(1)
			} else if !errors.Is(err, sql.ErrNoRows) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if successful.Load() != 1 {
		t.Fatal("token must succeed exactly once")
	}
	if _, err := db.Exec(`UPDATE users SET reset_token='expired', reset_token_expiry=CURRENT_TIMESTAMP - INTERVAL '1 minute'`); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ConsumePasswordReset("expired", "wrong"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expired token accepted: %v", err)
	}
	if _, err := repo.IssuePasswordReset(id, "latest"); err != nil {
		t.Fatal(err)
	}
	if err := repo.RevokePasswordReset(id, "expired"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ConsumePasswordReset("latest", "latest-password"); err != nil {
		t.Fatal("cleanup revoked a newer token", err)
	}
}
