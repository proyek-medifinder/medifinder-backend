package domain

import (
	"time"

	"github.com/google/uuid"
)

type Resep struct {
	ID          uuid.UUID `db:"id" json:"id"`
	TransaksiID uuid.UUID `db:"transaksi_id" json:"transaksi_id"`
	FilePath    string    `db:"file_path" json:"file_path"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
