package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

const RoleAdminUUID = "22222222-2222-2222-2222-222222222222"

type UserRepository struct {
	DB *sqlx.DB
}

func (r *UserRepository) GetUserProfile(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, name, email, role_id, status, created_at, updated_at 
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.DB.Get(&user, query, id)
	return &user, err
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
	// Diubah ke updated_at karena last_login_at gak ada di tabel users
	query := `UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.DB.Exec(query, id)
	return err
}

func (r *UserRepository) FindAdmins(limit, offset int) ([]domain.User, int, error) {
	var list []domain.User
	var total int

	err := r.DB.Get(&total, `SELECT COUNT(*) FROM users WHERE role_id = $1 AND deleted_at IS NULL`, RoleAdminUUID)
	if err != nil {
		return nil, 0, err
	}

	err = r.DB.Select(&list, `
		SELECT id, name, email, role_id
		FROM users
		WHERE role_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, RoleAdminUUID, limit, offset)

	return list, total, err
}

func (r *UserRepository) UpdateAdmin(id uuid.UUID, name, email string) error {
	_, err := r.DB.Exec(`
		UPDATE users
		SET name=$1, email=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id=$3 AND role_id=$4 AND deleted_at IS NULL
	`, name, email, id, RoleAdminUUID)
	return err
}

func (r *UserRepository) DeleteAdmin(id uuid.UUID) error {
	// ================= INI INTI DARI SOFT DELETE =================
	// Kita gak pakai DELETE FROM, tapi nge-update deleted_at jadi waktu saat ini
	_, err := r.DB.Exec(`
		UPDATE users
		SET deleted_at = CURRENT_TIMESTAMP, status = 'deleted'
		WHERE id=$1 AND role_id=$2 AND deleted_at IS NULL
	`, id, RoleAdminUUID)
	return err
}

func (r *UserRepository) CreateAdmin(user *domain.User) error {
	_, err := r.DB.Exec(`
		INSERT INTO users (id, name, email, password, role_id)
		VALUES ($1, $2, $3, $4, $5)
	`, user.ID, user.Name, user.Email, user.Password, RoleAdminUUID)
	return err
}

// ================= FITUR BARU RESET PASSWORD =================

func (r *UserRepository) UpdateResetToken(email, token string, expiry time.Time) error {
	query := `UPDATE users SET reset_token = $1, reset_token_expiry = $2 WHERE email = $3 AND deleted_at IS NULL`
	_, err := r.DB.Exec(query, token, expiry, email)
	return err
}

func (r *UserRepository) FindByResetToken(token string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE reset_token = $1 AND deleted_at IS NULL`
	err := r.DB.Get(&user, query, token)
	return &user, err
}

func (r *UserRepository) ClearResetToken(id uuid.UUID, newPassword string) error {
	query := `UPDATE users SET password = $1, reset_token = NULL, reset_token_expiry = NULL WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.DB.Exec(query, newPassword, id)
	return err
}

// =============== FITUR GANTI PASSWORD ======================
func (r *UserRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, password FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := r.DB.Get(&user, query, id)
	return &user, err
}

func (r *UserRepository) UpdatePassword(id uuid.UUID, newPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.DB.Exec(query, newPassword, id)
	return err
}

// ================ FITUR VERIFIKASI ADMIN OLEH SUPERADMIN =================
func (r *UserRepository) FindPendingAdmins(limit, offset int) ([]domain.User, int, error) {
	var list []domain.User
	var total int

	err := r.DB.Get(&total, `SELECT COUNT(*) FROM users WHERE role_id = $1 AND status = 'pending' AND deleted_at IS NULL`, RoleAdminUUID)
	if err != nil {
		return nil, 0, err
	}

	err = r.DB.Select(&list, `
		SELECT id, name, email, role_id, status, created_at
		FROM users
		WHERE role_id = $1 AND status = 'pending' AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`, RoleAdminUUID, limit, offset)

	return list, total, err
}

func (r *UserRepository) VerifyAdmin(adminID, superAdminID uuid.UUID, action, notes string) error {
	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Biar aman kalo panic/error di tengah jalan

	// 1. Update status user
	_, err = tx.Exec(`UPDATE users SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND role_id = $3 AND deleted_at IS NULL`, action, adminID, RoleAdminUUID)
	if err != nil {
		return err
	}

	// 2. Insert ke tabel verification_logs (Audit Trail)
	_, err = tx.Exec(`
		INSERT INTO verification_logs (id, admin_id, superadmin_id, action, notes)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), adminID, superAdminID, action, notes)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// ================= FITUR SUPER ADMIN UBAH STATUS (FR-45) =================

func (r *UserRepository) UpdateAdminStatus(id uuid.UUID, status string) error {
	query := `UPDATE users SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND role_id = $3 AND deleted_at IS NULL`
	_, err := r.DB.Exec(query, status, id, RoleAdminUUID)
	return err
}
