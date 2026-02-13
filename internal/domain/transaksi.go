package domain

import (
	"time"

	"github.com/google/uuid"
)

type Transaksi struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	ApotekID  uuid.UUID `db:"apotek_id" json:"apotek_id"`
	Total     int64     `db:"total" json:"total"`
	Status    string    `db:"status" json:"status"`
	ExpiredAt time.Time `db:"expired_at" json:"expired_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
