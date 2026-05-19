package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/dto"
)

const RoleAdminUUID = "22222222-2222-2222-2222-222222222222"

type UserRepository struct {
	DB *sqlx.DB
}

func (r *UserRepository) GetUserProfile(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `
		SELECT id, name, email, role_id, status, profile_picture, created_at, updated_at 
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
func (r *UserRepository) FindPendingAdmins(limit, offset int) ([]dto.PendingAdminResponse, int, error) {
	var list []dto.PendingAdminResponse
	var total int

	err := r.DB.Get(&total, `
		SELECT COUNT(*) 
		FROM users u
		JOIN admin_applications a ON u.id = a.user_id
		WHERE u.role_id = $1 AND u.status = 'pending' AND u.deleted_at IS NULL
	`, RoleAdminUUID)
	if err != nil {
		return nil, 0, err
	}

	// [UBAH] Lakukan JOIN dan pilih kolom-kolom yang ada di struct PendingAdminResponse
	err = r.DB.Select(&list, `
		SELECT 
			u.id AS user_id, u.name, u.email, u.role_id, u.status, u.created_at,
			a.id AS app_id, a.nama_apotek, a.alamat, a.latitude, a.longitude, a.phone_number, a.deskripsi, a.photo_url
		FROM users u
		JOIN admin_applications a ON u.id = a.user_id
		WHERE u.role_id = $1 AND u.status = 'pending' AND u.deleted_at IS NULL
		ORDER BY u.created_at ASC
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

// ================= FITUR ADMIN APOTEK REGISTER =================

// 1. Transaction untuk Registrasi Admin
func (r *UserRepository) RegisterAdminTx(user *domain.User, app *domain.AdminApplication) error {
	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Otomatis rollback kalau ada error sebelum di-commit

	// Insert ke tabel users
	_, err = tx.NamedExec(`
		INSERT INTO users (id, name, email, password, role_id, status, created_at, updated_at)
		VALUES (:id, :name, :email, :password, :role_id, :status, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, user)
	if err != nil {
		return err
	}

	// Insert ke tabel admin_applications
	_, err = tx.NamedExec(`
		INSERT INTO admin_applications (id, user_id, nama_apotek, alamat, latitude, longitude, phone_number, deskripsi, status, submitted_at)
		VALUES (:id, :user_id, :nama_apotek, :alamat, :latitude, :longitude, :phone_number, :deskripsi, :status, CURRENT_TIMESTAMP)
	`, app)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) ProcessAdminVerificationTx(adminID, superAdminID uuid.UUID, action, reason string) error {
	tx, err := r.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var app domain.AdminApplication
	err = tx.Get(&app, `SELECT * FROM admin_applications WHERE user_id = $1 ORDER BY submitted_at DESC LIMIT 1`, adminID)
	if err != nil {
		return err
	}

	userStatus := "rejected"
	appStatus := "REJECTED"
	var rejectionReason *string

	if action == "approved" {
		userStatus = "active"
		appStatus = "APPROVED"

		_, err = tx.Exec(`
			INSERT INTO apotek (id, admin_id, nama, alamat, latitude, longitude, phone_number, deskripsi, photo_url, verification_status, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'APPROVED', CURRENT_TIMESTAMP)
		`, uuid.New(), adminID, app.NamaApotek, app.Alamat, app.Latitude, app.Longitude, app.PhoneNumber, app.Deskripsi, app.PhotoURL)

		if err != nil {
			return err
		}
	} else {
		rejectionReason = &reason
	}

	_, err = tx.Exec(`
		UPDATE admin_applications 
		SET status = $1, rejection_reason = $2, reviewed_at = CURRENT_TIMESTAMP, reviewed_by = $3
		WHERE id = $4
	`, appStatus, rejectionReason, superAdminID, app.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE users SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND deleted_at IS NULL`, userStatus, adminID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO verification_logs (id, admin_id, superadmin_id, action, notes)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), adminID, superAdminID, action, reason)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) GetAdminApplicationByUserID(userID uuid.UUID) (*domain.AdminApplication, error) {
	var app domain.AdminApplication
	query := `
		SELECT id, user_id, nama_apotek, alamat, latitude, longitude, phone_number, deskripsi, status, submitted_at
		FROM admin_applications 
		WHERE user_id = $1 
		ORDER BY submitted_at DESC 
		LIMIT 1
	`
	err := r.DB.Get(&app, query, userID)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *UserRepository) GetAuthDataByID(userID uuid.UUID) (string, string, error) {
	var user struct {
		Email    string `db:"email"`
		Password string `db:"password"`
	}
	// Asumsi lu pake sqlx
	err := r.DB.Get(&user, "SELECT email, password FROM users WHERE id = $1", userID)
	return user.Email, user.Password, err
}

// ================= FITUR GOOGLE LOGIN =================
func (r *UserRepository) UpdateGoogleID(userID uuid.UUID, googleID string) error {
	query := `UPDATE users SET google_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.DB.Exec(query, googleID, userID)
	return err
}