package domain

import (
	"time"

	"github.com/google/uuid"
)

type AdminApplication struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	UserID          uuid.UUID  `db:"user_id" json:"user_id"`
	NamaApotek      string     `db:"nama_apotek" json:"nama_apotek"`
	Alamat          string     `db:"alamat" json:"alamat"`
	Latitude        float64    `db:"latitude" json:"latitude"`
	Longitude       float64    `db:"longitude" json:"longitude"`
	PhoneNumber     string     `db:"phone_number" json:"phone_number"`
	Deskripsi       string     `db:"deskripsi" json:"deskripsi"`
	Status          string     `db:"status" json:"status"` // PENDING, APPROVED, REJECTED
	RejectionReason *string    `db:"rejection_reason" json:"rejection_reason,omitempty"`
	SubmittedAt     time.Time  `db:"submitted_at" json:"submitted_at"`
	ReviewedAt      *time.Time `db:"reviewed_at" json:"reviewed_at,omitempty"`
	ReviewedBy      *uuid.UUID `db:"reviewed_by" json:"reviewed_by,omitempty"`
}
