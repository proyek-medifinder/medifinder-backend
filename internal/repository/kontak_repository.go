package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type KontakRepository struct {
	DB *sqlx.DB
}

func (r *KontakRepository) Create(kontak *domain.Kontak) error {
	query := `
		INSERT INTO kontak (id, nama, email, subjek, pesan, status)
		VALUES (:id, :nama, :email, :subjek, :pesan, :status)
	`
	_, err := r.DB.NamedExec(query, kontak)
	return err
}

func (r *KontakRepository) FindAll(limit, offset int) ([]domain.Kontak, error) {
	var list []domain.Kontak
	query := `
		SELECT id, nama, email, subjek, pesan, status, created_at 
		FROM kontak 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`
	err := r.DB.Select(&list, query, limit, offset)
	return list, err
}

func (r *KontakRepository) UpdateStatus(id, status string) error {
	query := `UPDATE kontak SET status = $1 WHERE id = $2`
	_, err := r.DB.Exec(query, status, id)
	return err
}

func (r *KontakRepository) FindByID(id string) (*domain.Kontak, error) {
    var kontak domain.Kontak
    query := `
        SELECT id, nama, email, subjek, pesan, status, created_at 
        FROM kontak 
        WHERE id = $1
    `
    err := r.DB.Get(&kontak, query, id)
    if err != nil {
        return nil, err
    }
    return &kontak, nil
}