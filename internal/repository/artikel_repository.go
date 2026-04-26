package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type ArtikelRepository struct {
	DB *sqlx.DB
}

func (r *ArtikelRepository) Create(artikel *domain.Artikel) error {
	query := `
	INSERT INTO artikel (id, superadmin_id, judul, slug, konten, kategori, image_url, status, source)
	VALUES (:id, :superadmin_id, :judul, :slug, :konten, :kategori, :image_url, :status, :source)
	ON CONFLICT (slug) DO NOTHING -- Biar ga duplikat kalau narik API
	`
	_, err := r.DB.NamedExec(query, artikel)
	return err
}

func (r *ArtikelRepository) GetPublished(limit, offset int) ([]domain.Artikel, error) {
	var list []domain.Artikel
	query := `
		SELECT id, judul, slug, konten, kategori, image_url, source, created_at 
		FROM artikel 
		WHERE status = 'PUBLISHED' 
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`
	err := r.DB.Select(&list, query, limit, offset)
	return list, err
}

// ================= TAMBAHAN BUAT SUPER ADMIN =================

func (r *ArtikelRepository) FindByID(id string) (*domain.Artikel, error) {
	var artikel domain.Artikel
	query := `SELECT * FROM artikel WHERE id = $1`
	err := r.DB.Get(&artikel, query, id)
	return &artikel, err
}

func (r *ArtikelRepository) Update(artikel *domain.Artikel) error {
	query := `
		UPDATE artikel 
		SET judul = :judul, slug = :slug, konten = :konten, kategori = :kategori, 
		    image_url = :image_url, status = :status, updated_at = CURRENT_TIMESTAMP
		WHERE id = :id
	`
	_, err := r.DB.NamedExec(query, artikel)
	return err
}

func (r *ArtikelRepository) Delete(id string) error {
	query := `DELETE FROM artikel WHERE id = $1`
	_, err := r.DB.Exec(query, id)
	return err
}
