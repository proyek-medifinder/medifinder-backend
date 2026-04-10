package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type ApotekRepository struct {
	DB *sqlx.DB
}

func (r *ApotekRepository) Create(apotek *domain.Apotek) error {
	query := `
	INSERT INTO apotek (id, admin_id, nama, alamat, latitude, longitude, jam_buka, jam_tutup)
	VALUES (:id, :admin_id, :nama, :alamat, :latitude, :longitude, :jam_buka, :jam_tutup)
	`
	_, err := r.DB.NamedExec(query, apotek)
	return err
}

func (r *ApotekRepository) FindByAdmin(adminID uuid.UUID) (*domain.Apotek, error) {
	var apotek domain.Apotek

	query := `
		SELECT id, admin_id, nama, alamat, latitude, longitude, jam_buka, jam_tutup, 
		       phone_number, deskripsi, verification_status, rejection_reason, created_at 
		FROM apotek 
		WHERE admin_id = $1
	`

	err := r.DB.Get(&apotek, query, adminID)
	if err != nil {
		return nil, err
	}
	return &apotek, nil
}

func (r *ObatRepository) FindByApotekPaginated(apotekID string, name string, limit int, offset int) ([]domain.Obat, error) {
	var obat []domain.Obat
	var query string
	var args []interface{}

	if name != "" {
		query = `
			SELECT id, apotek_id, nama, stok, harga 
			FROM obat 
			WHERE apotek_id=$1 AND (nama ILIKE $2 OR similarity(nama, $3) > 0.3)
			ORDER BY similarity(nama, $3) DESC
			LIMIT $4 OFFSET $5
		`
		args = []interface{}{apotekID, "%" + name + "%", name, limit, offset}
	} else {
		query = `SELECT id, apotek_id, nama, stok, harga FROM obat WHERE apotek_id=$1 LIMIT $2 OFFSET $3`
		args = []interface{}{apotekID, limit, offset}
	}

	err := r.DB.Select(&obat, query, args...)
	return obat, err
}

func (r *ObatRepository) CountByApotek(apotekID string, name string) (int, error) {
	var count int
	var query string
	var args []interface{}

	if name != "" {
		query = `
			SELECT COUNT(id) 
			FROM obat 
			WHERE apotek_id=$1 AND (nama ILIKE $2 OR similarity(nama, $3) > 0.3)
		`
		args = []interface{}{apotekID, "%" + name + "%", name}
	} else {
		query = `SELECT COUNT(id) FROM obat WHERE apotek_id=$1`
		args = []interface{}{apotekID}
	}

	err := r.DB.Get(&count, query, args...)
	return count, err
}

func (r *ApotekRepository) Update(apotek *domain.Apotek) error {
	query := `
	UPDATE apotek 
	SET nama=:nama, alamat=:alamat, latitude=:latitude, longitude=:longitude, 
	    jam_buka=:jam_buka, jam_tutup=:jam_tutup, phone_number=:phone_number, deskripsi=:deskripsi
	WHERE id=:id
	`
	_, err := r.DB.NamedExec(query, apotek)
	return err
}

// Fitur Filter Jarak + Jam Operasional Terbuka
func (r *ApotekRepository) FindNearby(
	lat, lng float64,
	radius float64,
	limit, offset int,
	currentTime string,
) ([]domain.Apotek, int, error) {

	var list []domain.Apotek
	var total int

	baseQuery := `
	FROM apotek
	WHERE (
		6371 * acos(
			cos(radians($1)) *
			cos(radians(latitude)) *
			cos(radians(longitude) - radians($2)) +
			sin(radians($1)) *
			sin(radians(latitude))
		)
	) <= $3
	AND (
		-- Tambahin ::time di setiap jam_buka dan jam_tutup
		(jam_buka::time <= jam_tutup::time AND $4::time >= jam_buka::time AND $4::time <= jam_tutup::time)
		OR
		(jam_buka::time > jam_tutup::time AND ($4::time >= jam_buka::time OR $4::time <= jam_tutup::time))
	)
	`

	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.DB.Get(&total, countQuery, lat, lng, radius, currentTime)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
	SELECT id, nama, alamat, latitude, longitude, jam_buka, jam_tutup,
	(
		6371 * acos(
			cos(radians($1)) *
			cos(radians(latitude)) *
			cos(radians(longitude) - radians($2)) +
			sin(radians($1)) *
			sin(radians(latitude))
		)
	) AS distance
	` + baseQuery + `
	ORDER BY distance ASC
	LIMIT $5 OFFSET $6
	`

	err = r.DB.Select(&list, dataQuery,
		lat, lng, radius, currentTime,
		limit, offset,
	)

	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
