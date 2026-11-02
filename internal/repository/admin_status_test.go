package repository

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"os"
	"testing"
)

func TestAdminStatusLifecycle(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; requires disposable PostgreSQL")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TEMP TABLE users (id uuid PRIMARY KEY, name text, email text, role_id uuid, status text,
        deleted_at timestamptz, created_at timestamptz DEFAULT CURRENT_TIMESTAMP, updated_at timestamptz DEFAULT CURRENT_TIMESTAMP);
        CREATE TEMP TABLE apotek (id uuid, admin_id uuid, nama text, created_at timestamptz DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		t.Fatal(err)
	}
	activeID, pendingID := uuid.New(), uuid.New()
	_, err = db.Exec(`INSERT INTO users(id,name,email,role_id,status) VALUES
        ($1,'Active admin','admin@example.invalid',$3,'active'),($2,'Pending admin','pending@example.invalid',$3,'pending');
        `, activeID, pendingID, RoleAdminUUID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO apotek(id,admin_id,nama) VALUES ($1,$2,'Apotek Uji')`, uuid.New(), activeID); err != nil {
		t.Fatal(err)
	}
	repo := &UserRepository{DB: db}
	admins, total, err := repo.FindAdmins(50, 0)
	if err != nil || total != 2 {
		t.Fatalf("admin listing: %v", err)
	}
	for _, admin := range admins {
		if admin.Status == "" || admin.CreatedAt.Year() < 2000 {
			t.Fatal("status or registration date missing from API data")
		}
		if admin.ID == activeID && admin.PharmacyName != "Apotek Uji" {
			t.Fatal("wrong pharmacy name")
		}
	}
	if _, changed, err := repo.UpdateAdminStatus(activeID, "active"); err != nil || changed {
		t.Fatal("repeating status must not trigger notification")
	}
	if _, changed, err := repo.UpdateAdminStatus(activeID, "suspended"); err != nil || !changed {
		t.Fatal("deactivation failed", err)
	}
	if active, err := repo.IsActiveAdmin(activeID.String()); err != nil || active {
		t.Fatal("deactivated admin still active")
	}
	if _, changed, err := repo.UpdateAdminStatus(activeID, "active"); err != nil || !changed {
		t.Fatal("reactivation failed", err)
	}
	if active, err := repo.IsActiveAdmin(activeID.String()); err != nil || !active {
		t.Fatal("reactivated admin blocked")
	}
	if _, _, err := repo.UpdateAdminStatus(pendingID, "active"); err == nil {
		t.Fatal("activation bypassed registration verification")
	}
	deleted, err := repo.DeleteAdmin(activeID)
	if err != nil {
		t.Fatal(err)
	}
	if deleted.ID != activeID || deleted.Email != "admin@example.invalid" || deleted.Name != "Active admin" {
		t.Fatal("deletion did not return the notification recipient")
	}
	if _, err := repo.DeleteAdmin(activeID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("repeated deletion must not return another notification recipient", err)
	}
}
