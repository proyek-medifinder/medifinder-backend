package domain

import (
	"time"

	"github.com/google/uuid"
)

type Kontak struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Nama      string    `db:"nama" json:"nama"`
	Email     string    `db:"email" json:"email"`
	Subjek    string    `db:"subjek" json:"subjek"`
	Pesan     string    `db:"pesan" json:"pesan"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
