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
	JamBuka   string    `db:"jam_buka" json:"jam_buka"`
	JamTutup  string    `db:"jam_tutup" json:"jam_tutup"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
