package domain

import (
	"time"

	"github.com/google/uuid"
)

type Resep struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	FilePath  string    `db:"file_path" json:"file_path"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
