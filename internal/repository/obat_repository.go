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

// ================= CREATE =================
func (r *ObatRepository) Create(obat *domain.Obat) error {
	query := `
	INSERT INTO obat (id, apotek_id, nama, stok, reserved_stock, harga)
	VALUES (:id, :apotek_id, :nama, :stok, :reserved_stock, :harga)
	`
	_, err := r.DB.NamedExec(query, obat)
	return err
}

// ================= FIND BY APOTEK =================
func (r *ObatRepository) FindByApotek(apotekID string) ([]domain.Obat, error) {
	var obat []domain.Obat

	query := `
	SELECT id, apotek_id, nama, stok, reserved_stock, harga::INT
	FROM obat
	WHERE apotek_id = $1
	`

	err := r.DB.Select(&obat, query, apotekID)
	return obat, err
}

// ================= FIND BY ID =================
func (r *ObatRepository) FindByID(id string) (*domain.Obat, error) {
	var obat domain.Obat

	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}

	query := `
	SELECT id, apotek_id, nama, stok, reserved_stock, harga::INT
	FROM obat
	WHERE id = $1
	`

	err = r.DB.Get(&obat, query, parsedID)
	if err != nil {
		return nil, err
	}

	return &obat, nil
}

// ================= UPDATE =================
func (r *ObatRepository) Update(obat *domain.Obat) error {
	query := `
	UPDATE obat 
	SET nama = :nama,
		stok = :stok,
		reserved_stock = :reserved_stock,
		harga = :harga
	WHERE id = :id
	`
	_, err := r.DB.NamedExec(query, obat)
	return err
}

// ================= DELETE =================
func (r *ObatRepository) Delete(id string) error {
	query := `DELETE FROM obat WHERE id = $1`

	_, err := r.DB.Exec(query, id)
	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return errors.New("obat tidak bisa dihapus karena sudah tercatat di transaksi. Ubah stok jadi 0 saja")
		}
		return err
	}

	return nil
}

type ApotekWithMedicineStock struct {
	ApotekID  string  `db:"apotek_id"`
	Nama      string  `db:"nama"`
	Alamat    string  `db:"alamat"`
	Latitude  float64 `db:"latitude"`   
	Longitude float64 `db:"longitude"` 
	Harga     float64 `db:"harga"`
	Stok      int     `db:"stok"`
	Distance  float64 `db:"distance"`
	JamBuka   *string `db:"jam_buka"`
	JamTutup  *string `db:"jam_tutup"`
}

// Cari apotek berstok dari jarak paling dekat ke paling jauh
func (r *ObatRepository) FindPharmaciesWithStock(lat, lng float64, medicineName string) ([]ApotekWithMedicineStock, error) {
	var list []ApotekWithMedicineStock

	query := `
	SELECT 
		a.id::TEXT AS apotek_id, 
		a.nama, 
		a.alamat, 
		a.latitude,
		a.longitude,
		COALESCE(a.jam_buka::TEXT, '') AS jam_buka,
		COALESCE(a.jam_tutup::TEXT, '') AS jam_tutup,
		o.harga::FLOAT AS harga, 
		(COALESCE(o.stok, 0) - COALESCE(o.reserved_stock, 0))::INT AS stok,
		(6371 * acos(
			LEAST(1.0, GREATEST(-1.0, 
				cos(radians($1)) * cos(radians(a.latitude)) * 
				cos(radians(a.longitude) - radians($2)) + 
				sin(radians($1)) * sin(radians(a.latitude))
			))
		)) AS distance
	FROM apotek a
	JOIN obat o ON o.apotek_id = a.id
	WHERE o.nama ILIKE $3
	  AND COALESCE(o.stok, 0) > 0
	  AND UPPER(a.verification_status) = 'APPROVED'
	ORDER BY distance ASC
	LIMIT 5;
	`

	searchTerm := "%" + strings.TrimSpace(medicineName) + "%"
	err := r.DB.Select(&list, query, lat, lng, searchTerm)
	return list, err
}
