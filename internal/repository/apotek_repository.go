package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type ApotekRepository struct {
	DB *sqlx.DB
}

func (r *ApotekRepository) Create(apotek *domain.Apotek) error {

	query := `
        INSERT INTO apotek (id, admin_id, nama, alamat, latitude, longitude, jam_buka, jam_tutup, phone_number, deskripsi, photo_url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `
	// Ambil isi pointer photo_url biar aman
	var url interface{}
	if apotek.PhotoURL != nil {
		url = *apotek.PhotoURL
	}

	_, err := r.DB.Exec(query,
		apotek.ID, apotek.AdminID, apotek.Nama, apotek.Alamat,
		apotek.Latitude, apotek.Longitude, apotek.JamBuka, apotek.JamTutup,
		apotek.PhoneNumber, apotek.Deskripsi, url)

	return err
}

func (r *ApotekRepository) FindByAdmin(adminID uuid.UUID) (*domain.Apotek, error) {
	var apotek domain.Apotek

	query := `
		SELECT id, admin_id, nama, alamat, latitude, longitude, jam_buka, jam_tutup, 
		       phone_number, deskripsi, verification_status, rejection_reason, created_at, photo_url 
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
	    jam_buka=:jam_buka, jam_tutup=:jam_tutup, phone_number=:phone_number, deskripsi=:deskripsi, photo_url=:photo_url
	WHERE id=:id
	`
	_, err := r.DB.NamedExec(query, apotek)
	return err
}

func (r *ApotekRepository) FindNearby(
	lat, lng float64,
	radius float64,
	limit, offset int,
	currentTime string,
) ([]domain.Apotek, int, error) {

	list := []domain.Apotek{}
	var total int

	// 1. Kondisi Filter Jarak (Opsional kalau lat/lng tidak 0)
	distanceCondition := "TRUE" // Default kalau ga ada koordinat
	distanceSelect := "0 AS distance"
	if lat != 0 && lng != 0 {
		distanceCondition = `(
			6371 * acos(
				cos(radians($1)) *
				cos(radians(latitude)) *
				cos(radians(longitude) - radians($2)) +
				sin(radians($1)) *
				sin(radians(latitude))
			)
		) <= $3`
		distanceSelect = `(
			6371 * acos(
				cos(radians($1)) *
				cos(radians(latitude)) *
				cos(radians(longitude) - radians($2)) +
				sin(radians($1)) *
				sin(radians(latitude))
			)
		) AS distance`
	}

	// 2. Kondisi Filter Jam (Handle NULL supaya tetap muncul)
	// Kita tambah "jam_buka IS NULL" supaya data lu yg NULL ga ilang
	timeCondition := `(
		(jam_buka IS NULL OR jam_tutup IS NULL)
		OR
		(jam_buka::time <= jam_tutup::time AND $4::time >= jam_buka::time AND $4::time <= jam_tutup::time)
		OR
		(jam_buka::time > jam_tutup::time AND ($4::time >= jam_buka::time OR $4::time <= jam_tutup::time))
	)`

	baseQuery := " FROM apotek WHERE " + distanceCondition + " AND " + timeCondition

	// 3. Eksekusi Count Query
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.DB.Get(&total, countQuery, lat, lng, radius, currentTime)
	if err != nil {
		return nil, 0, err
	}

	// 4. Eksekusi Data Query
	dataQuery := "SELECT id, nama, alamat, latitude, longitude, jam_buka, jam_tutup, photo_url, " +
		distanceSelect + baseQuery +
		" ORDER BY distance ASC LIMIT $5 OFFSET $6"

	err = r.DB.Select(&list, dataQuery,
		lat, lng, radius, currentTime,
		limit, offset,
	)

	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *ApotekRepository) GetByID(id string) (domain.Apotek, error) {
	var apotek domain.Apotek

	// 1. Ambil data apoteknya dulu
	queryApotek := `SELECT id, admin_id, nama, alamat, latitude, longitude, phone_number, deskripsi, jam_buka, jam_tutup, photo_url, verification_status, created_at FROM apotek WHERE id = $1`
	err := r.DB.Get(&apotek, queryApotek, id)
	if err != nil {
		return apotek, err
	}

	// 2. Inisialisasi slice dengan array kosong
	obats := []domain.Obat{}

	// 3. FIX: Tambahin ::INT di kolom harga biar gak bentrok sama int64 di struct Go
	queryObat := `SELECT id, apotek_id, nama, stok, harga::INT FROM obat WHERE apotek_id = $1`
	err = r.DB.Select(&obats, queryObat, id)
	if err != nil {
		fmt.Printf("ERROR AMBIL OBAT: %v\n", err)
		return apotek, nil
	}

	apotek.Obats = obats

	return apotek, nil
}

func (r *ApotekRepository) FindAllWithCount(limit, offset int) ([]domain.Apotek, int, error) {
	var list []domain.Apotek
	var total int

	// 1. Hitung total semua apotek
	err := r.DB.Get(&total, "SELECT COUNT(*) FROM apotek")
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, admin_id, nama, alamat, latitude, longitude, phone_number, photo_url, verification_status, created_at 
		FROM apotek 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`
	err = r.DB.Select(&list, query, limit, offset)

	return list, total, err
}
