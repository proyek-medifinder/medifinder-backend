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
	var url string
	if apotek.PhotoURL != nil {
		url = *apotek.PhotoURL
	}

	_, err := r.DB.Exec(query,
		apotek.ID, apotek.AdminID, apotek.Nama, apotek.Alamat,
		apotek.Latitude, apotek.Longitude, apotek.JamBuka, apotek.JamTutup,
		apotek.PhoneNumber, apotek.Deskripsi, url) // <--- Pake url (string)

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
    SET nama=$1, alamat=$2, latitude=$3, longitude=$4, 
        jam_buka=$5, jam_tutup=$6, phone_number=$7, deskripsi=$8, photo_url=$9
    WHERE id=$10
    `
	// Ambil nilai dari pointer, kalau nil kasih string kosong
	var pUrl, phone, desk, buka, tutup string
	if apotek.PhotoURL != nil {
		pUrl = *apotek.PhotoURL
	}
	if apotek.PhoneNumber != nil {
		phone = *apotek.PhoneNumber
	}
	if apotek.Deskripsi != nil {
		desk = *apotek.Deskripsi
	}
	if apotek.JamBuka != nil {
		buka = *apotek.JamBuka
	}
	if apotek.JamTutup != nil {
		tutup = *apotek.JamTutup
	}

	_, err := r.DB.Exec(query,
		apotek.Nama, apotek.Alamat, apotek.Latitude, apotek.Longitude,
		buka, tutup, phone, desk, pUrl, apotek.ID)

	if err != nil {
		fmt.Printf("ERROR UPDATE DB: %v\n", err) // INI BAKAL KELUAR DI RAILWAY!
		return err
	}
	return nil
}

func (r *ApotekRepository) FindNearby(
	lat, lng float64,
	radius float64,
	limit, offset int,
	currentTime string,
) ([]domain.Apotek, int, error) {

	list := []domain.Apotek{}
	var total int

	// Array untuk nampung parameter secara dinamis
	var args []interface{}

	// 1. Kondisi Filter Jarak (Dinamic)
	distanceCondition := "TRUE"
	distanceSelect := "0 AS distance" // Default kalau ga ada koordinat

	if lat != 0 && lng != 0 {
		// Kalau koordinat ada, kita pake index $1, $2, $3
		distanceCondition = `(
			6371 * acos(
				cos(radians($1)) * cos(radians(latitude)) * cos(radians(longitude) - radians($2)) +
				sin(radians($1)) * sin(radians(latitude))
			)
		) <= $3`
		distanceSelect = `(
			6371 * acos(
				cos(radians($1)) * cos(radians(latitude)) * cos(radians(longitude) - radians($2)) +
				sin(radians($1)) * sin(radians(latitude))
			)
		) AS distance`
		args = append(args, lat, lng, radius)
	}

	// Tentukan index buat currentTime (bisa $4 kalau ada koordinat, bisa $1 kalau ga ada)
	timeParam := fmt.Sprintf("$%d", len(args)+1)
	args = append(args, currentTime)

	// 2. Kondisi Filter Jam (SUPER AMAN pakai NULLIF)
	// Kita ubah jam_buka == "" jadi NULL biar Postgres ga ngamuk saat nge-cast ::time
	timeCondition := fmt.Sprintf(`(
		(NULLIF(jam_buka, '') IS NULL OR NULLIF(jam_tutup, '') IS NULL)
		OR
		(NULLIF(jam_buka, '')::time <= NULLIF(jam_tutup, '')::time AND %[1]s::time >= NULLIF(jam_buka, '')::time AND %[1]s::time <= NULLIF(jam_tutup, '')::time)
		OR
		(NULLIF(jam_buka, '')::time > NULLIF(jam_tutup, '')::time AND (%[1]s::time >= NULLIF(jam_buka, '')::time OR %[1]s::time <= NULLIF(jam_tutup, '')::time))
	)`, timeParam)

	baseQuery := " FROM apotek WHERE " + distanceCondition + " AND " + timeCondition

	// 3. Eksekusi Count
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.DB.Get(&total, countQuery, args...)
	if err != nil {
		fmt.Printf("ERROR COUNT NEARBY: %v\n", err)
		return nil, 0, err
	}

	// Tambahin limit & offset ke argument list
	limitParam := fmt.Sprintf("$%d", len(args)+1)
	offsetParam := fmt.Sprintf("$%d", len(args)+2)
	args = append(args, limit, offset)

	// 4. Eksekusi Data
	dataQuery := "SELECT id, admin_id, nama, alamat, latitude, longitude, jam_buka, jam_tutup, phone_number, deskripsi, photo_url, verification_status, created_at, " +
		distanceSelect + baseQuery +
		" ORDER BY distance ASC LIMIT " + limitParam + " OFFSET " + offsetParam

	err = r.DB.Select(&list, dataQuery, args...)
	if err != nil {
		fmt.Printf("ERROR SELECT NEARBY: %v\n", err)
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

func (r *ApotekRepository) UpdatePhoto(adminID uuid.UUID, url string) error {
	// Pake query sederhana, update langsung ke tabel apotek berdasarkan admin_id
	query := `UPDATE apotek SET photo_url = $1 WHERE admin_id = $2`

	fmt.Printf("DEBUG FINAL: SQL UPDATE apotek SET photo_url='%s' WHERE admin_id='%s'\n", url, adminID)

	result, err := r.DB.Exec(query, url, adminID)
	if err != nil {
		fmt.Printf("DEBUG: Gagal update foto di DB: %v\n", err)
		return err
	}

	rows, _ := result.RowsAffected()
	fmt.Printf("DEBUG: Berhasil update %d baris di DB\n", rows)
	return nil
}
