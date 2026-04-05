package repository

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type ObatRepository struct {
	DB *sqlx.DB
}

func (r *ObatRepository) Create(obat *domain.Obat) error {
	query := `
	INSERT INTO obat (id, apotek_id, nama, stok, harga)
	VALUES (:id, :apotek_id, :nama, :stok, :harga)
	`
	_, err := r.DB.NamedExec(query, obat)
	return err
}

func (r *ObatRepository) FindByApotek(apotekID string) ([]domain.Obat, error) {
	var obat []domain.Obat
	query := `SELECT id, apotek_id, nama, stok, harga::INT FROM obat WHERE apotek_id=$1`
	err := r.DB.Select(&obat, query, apotekID)
	return obat, err
}

func (r *ObatRepository) FindByID(id string) (*domain.Obat, error) {
	var obat domain.Obat

	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}

	// Tambahin ::INT di kolom harga biar gak bentrok sama int64 di Go
	query := `SELECT id, apotek_id, nama, stok, harga::INT FROM obat WHERE id = $1`
	err = r.DB.Get(&obat, query, parsedID)

	if err != nil {
		return nil, err
	}

	return &obat, nil
}

func (r *ObatRepository) Update(obat *domain.Obat) error {
	query := `
	UPDATE obat 
	SET nama=:nama, stok=:stok, harga=:harga
	WHERE id=:id
	`
	_, err := r.DB.NamedExec(query, obat)
	return err
}

func (r *ObatRepository) Delete(id string) error {
	query := `DELETE FROM obat WHERE id=$1`
	_, err := r.DB.Exec(query, id)

	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return errors.New("obat tidak bisa dihapus karena sudah tercatat di riwayat transaksi pengguna. Silakan update stok menjadi 0 saja")
		}
		return err
	}

	return nil
}
