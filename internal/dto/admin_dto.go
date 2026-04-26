package dto

import (
	"github.com/google/uuid"
)

type VerifyAdminRequest struct {
	AdminID string `json:"admin_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Action  string `json:"action" binding:"required" example:"approved"`
	Notes   string `json:"notes" example:"Dokumen lengkap dan valid"`
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
	ID          uuid.UUID `db:"id" json:"id"`
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	NamaApotek  string    `db:"nama_apotek" json:"nama_apotek"`
	Alamat      string    `db:"alamat" json:"alamat"`
	Latitude    float64   `db:"latitude" json:"latitude"`
	Longitude   float64   `db:"longitude" json:"longitude"`
	PhoneNumber string    `db:"phone_number" json:"phone_number"`
	Deskripsi   *string   `db:"deskripsi" json:"deskripsi,omitempty"`
	PhotoURL    *string   `db:"photo_url" json:"photo_url,omitempty"`
	Status      string    `db:"status" json:"status"`
}
