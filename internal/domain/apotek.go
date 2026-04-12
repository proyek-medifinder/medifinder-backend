package domain

import (
	"time"

	"github.com/google/uuid"
)

type Apotek struct {
	ID        uuid.UUID `db:"id" json:"id"`
	AdminID   uuid.UUID `db:"admin_id" json:"admin_id"`
	Nama      string    `db:"nama" json:"nama"`
	Alamat    string    `db:"alamat" json:"alamat"`
	Latitude  float64   `db:"latitude" json:"latitude"`
	Longitude float64   `db:"longitude" json:"longitude"`

	// UBAH 3 FIELD INI JADI POINTER (*string) AGAR BISA MENERIMA NULL DARI DATABASE
	PhoneNumber *string `db:"phone_number" json:"phone_number"`
	Deskripsi   *string `db:"deskripsi" json:"deskripsi,omitempty"`
	JamBuka     *string `db:"jam_buka" json:"jam_buka"`
	JamTutup    *string `db:"jam_tutup" json:"jam_tutup"`

	VerificationStatus string    `db:"verification_status" json:"verification_status"`
	RejectionReason    *string   `db:"rejection_reason" json:"rejection_reason,omitempty"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	Distance           *float64  `db:"distance" json:"distance,omitempty"`
}
