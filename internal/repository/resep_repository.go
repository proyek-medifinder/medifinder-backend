package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type ResepRepository struct {
	DB *sqlx.DB
}

func (r *ResepRepository) Create(resep *domain.Resep) error {
	query := `
	INSERT INTO resep (id, transaksi_id, file_path)
	VALUES ($1, $2, $3)
	`
	_, err := r.DB.Exec(query,
		resep.ID,
		resep.TransaksiID,
		resep.FilePath,
	)
	return err
}

func (r *ResepRepository) UpdateStatus(id uuid.UUID, status string) error {

	_, err := r.DB.Exec(`
		UPDATE resep
		SET status = $1
		WHERE id = $2
	`, status, id)

	return err
}

func (r *ResepRepository) FindAll(limit, offset int) ([]domain.Resep, int, error) {
	var list []domain.Resep
	var total int

	err := r.DB.Get(&total, `SELECT COUNT(*) FROM resep`)
	if err != nil {
		return nil, 0, err
	}

	err = r.DB.Select(&list, `
		SELECT id, transaksi_id, file_path, created_at
		FROM resep
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	return list, total, err
}
