package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type UserRepository struct {
	DB *sqlx.DB
}

func (r *UserRepository) Create(user *domain.User) error {
	query := `
	INSERT INTO users (id, name, email, password, role_id, google_id, status)
	VALUES (:id, :name, :email, :password, :role_id, :google_id, :status)
	`
	_, err := r.DB.NamedExec(query, user)
	return err
}

func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	query := `
	SELECT id, name, email, password, role_id, status 
	FROM users 
	WHERE LOWER(email)=LOWER($1) AND deleted_at IS NULL
	`
	err := r.DB.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.DB.Exec(query, id)
	return err
}

func (r *UserRepository) FindAdmins(limit, offset int) ([]domain.User, int, error) {
	var list []domain.User
	var total int

	err := r.DB.Get(&total, `SELECT COUNT(*) FROM users WHERE role_id = 2`)
	if err != nil {
		return nil, 0, err
	}

	err = r.DB.Select(&list, `
		SELECT id, name, email, role_id
		FROM users
		WHERE role_id = 2
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	return list, total, err
}

func (r *UserRepository) UpdateAdmin(id uuid.UUID, name, email string) error {
	_, err := r.DB.Exec(`
		UPDATE users
		SET name=$1, email=$2
		WHERE id=$3 AND role_id=2
	`, name, email, id)
	return err
}

func (r *UserRepository) DeleteAdmin(id uuid.UUID) error {
	_, err := r.DB.Exec(`
		DELETE FROM users
		WHERE id=$1 AND role_id=2
	`, id)
	return err
}

func (r *UserRepository) CreateAdmin(user *domain.User) error {
	_, err := r.DB.Exec(`
		INSERT INTO users (id, name, email, password, role_id)
		VALUES ($1, $2, $3, $4, 2)
	`, user.ID, user.Name, user.Email, user.Password)
	return err
}

// ================= FITUR BARU RESET PASSWORD =================

func (r *UserRepository) UpdateResetToken(email, token string, expiry time.Time) error {
	query := `UPDATE users SET reset_token = $1, reset_token_expiry = $2 WHERE email = $3`
	_, err := r.DB.Exec(query, token, expiry, email)
	return err
}

func (r *UserRepository) FindByResetToken(token string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE reset_token = $1`
	err := r.DB.Get(&user, query, token)
	return &user, err
}

func (r *UserRepository) ClearResetToken(id uuid.UUID, newPassword string) error {
	query := `UPDATE users SET password = $1, reset_token = NULL, reset_token_expiry = NULL WHERE id = $2`
	_, err := r.DB.Exec(query, newPassword, id)
	return err
}

// ================ FITUR VERIFIKASI ADMIN OLEH SUPERADMIN =================
func (r *UserRepository) FindPendingAdmins(limit, offset int) ([]domain.User, int, error) {
	var list []domain.User
	var total int

	err := r.DB.Get(&total, `SELECT COUNT(*) FROM users WHERE role_id = 2 AND status = 'pending' AND deleted_at IS NULL`)
	if err != nil {
		return nil, 0, err
	}

	err = r.DB.Select(&list, `
		SELECT id, name, email, role_id, status, created_at
		FROM users
		WHERE role_id = 2 AND status = 'pending' AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	return list, total, err
}

func (r *UserRepository) VerifyAdmin(adminID, superAdminID uuid.UUID, action, notes string) error {
	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}

	// 1. Update status user
	_, err = tx.Exec(`UPDATE users SET status = $1 WHERE id = $2 AND role_id = 2`, action, adminID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Insert ke tabel verification_logs (Audit Trail)
	_, err = tx.Exec(`
		INSERT INTO verification_logs (admin_id, superadmin_id, action, notes)
		VALUES ($1, $2, $3, $4)
	`, adminID, superAdminID, action, notes)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
