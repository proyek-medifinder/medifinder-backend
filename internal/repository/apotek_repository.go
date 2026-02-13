package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type ApotekRepository struct {
	DB *sqlx.DB
}

func (r *ApotekRepository) Create(apotek *domain.Apotek) error {
	query := `
	INSERT INTO apotek (id, admin_id, nama, alamat, latitude, longitude)
	VALUES (:id, :admin_id, :nama, :alamat, :latitude, :longitude)
	`
	_, err := r.DB.NamedExec(query, apotek)
	return err
}

func (r *ApotekRepository) FindByAdmin(adminID string) (*domain.Apotek, error) {
	var apotek domain.Apotek
	query := `SELECT id, admin_id, nama, alamat, latitude, longitude 
          FROM apotek 
          WHERE admin_id = $1::uuid`
	err := r.DB.Get(&apotek, query, adminID)
	if err != nil {
		return nil, err
	}
	return &apotek, nil
}

func (r *ObatRepository) FindByApotekPaginated(apotekID string, limit, offset int) ([]domain.Obat, error) {

	var obat []domain.Obat

	query := `
	SELECT id, apotek_id, nama, stok, harga
	FROM obat
	WHERE apotek_id=$1
	ORDER BY nama ASC
	LIMIT $2 OFFSET $3
	`

	err := r.DB.Select(&obat, query, apotekID, limit, offset)
	return obat, err
}

func (r *ObatRepository) CountByApotek(apotekID string) (int, error) {
	var total int
	query := `SELECT COUNT(*) FROM obat WHERE apotek_id=$1`
	err := r.DB.Get(&total, query, apotekID)
	return total, err
}

func (r *ApotekRepository) Update(apotek *domain.Apotek) error {
	query := `
	UPDATE apotek 
	SET nama=:nama, alamat=:alamat, latitude=:latitude, longitude=:longitude
	WHERE id=:id
	`
	_, err := r.DB.NamedExec(query, apotek)
	return err
}

func (r *ApotekRepository) FindNearby(
	lat, lng float64,
	radius float64,
	limit, offset int,
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
	`

	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.DB.Get(&total, countQuery, lat, lng, radius)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
	SELECT id, nama, alamat, latitude, longitude,
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
	LIMIT $4 OFFSET $5
	`

	err = r.DB.Select(&list, dataQuery,
		lat, lng, radius,
		limit, offset,
	)

	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
