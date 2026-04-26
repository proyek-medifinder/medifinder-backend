package domain

import (
	"time"

	"github.com/google/uuid"
)

type Artikel struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	SuperAdminID *uuid.UUID `db:"superadmin_id" json:"superadmin_id"`
	Judul        string     `db:"judul" json:"judul"`
	Slug         string     `db:"slug" json:"slug"`
	Konten       string     `db:"konten" json:"konten"`
	Kategori     string     `db:"kategori" json:"kategori"`
	ImageURL     *string    `db:"image_url" json:"image_url"`
	Status       string     `db:"status" json:"status"`
	Source       string     `db:"source" json:"source"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}
