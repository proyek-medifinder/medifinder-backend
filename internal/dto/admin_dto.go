package dto

import (
	"time"
	"github.com/google/uuid"
)

type VerifyAdminRequest struct {
    AdminID string `json:"admin_id" binding:"required"`
    Action  string `json:"action" binding:"required"` // "approved" atau "rejected"
    Reason  string `json:"reason"` // opsional
}

type ChangeAdminStatusRequest struct {
	Status string `json:"status" binding:"required" example:"suspended"`
}

type AdminRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateAdminRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type PendingAdminResponse struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	AppID       uuid.UUID `json:"app_id" db:"app_id"`
	NamaApotek  string    `json:"nama_apotek" db:"nama_apotek"`
	Alamat      string    `json:"alamat" db:"alamat"`
	Latitude    float64   `json:"latitude" db:"latitude"`
	Longitude   float64   `json:"longitude" db:"longitude"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Deskripsi   *string   `json:"deskripsi" db:"deskripsi"`
	PhotoURL    *string   `json:"photo_url" db:"photo_url"`
}
