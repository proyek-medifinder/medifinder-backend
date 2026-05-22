package repository

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type TransaksiRepository struct {
	DB *sqlx.DB
}

func (r *TransaksiRepository) FindByUser(
	userID string,
	status string,
	limit, offset int,
) ([]domain.Transaksi, int, error) {

	var list []domain.Transaksi
	var total int

	baseQuery := `
	FROM transaksi
	WHERE user_id = $1
	`

	args := []interface{}{uuid.MustParse(userID)}
	argIndex := 2

	if status != "" {
		baseQuery += " AND status = $" + strconv.Itoa(argIndex)
		args = append(args, status)
		argIndex++
	}

	// count total
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// data query yang udah diupdate (Hapus expired_at, tambah token & url) adsadasd
	dataQuery := `
	SELECT id, user_id, apotek_id, total_harga, status, snap_token, payment_url, created_at, updated_at
	` + baseQuery + `
	ORDER BY created_at DESC
	LIMIT $` + strconv.Itoa(argIndex) +
		` OFFSET $` + strconv.Itoa(argIndex+1)

	args = append(args, limit, offset)

	// Nah, variabel dataQuery dipake di sini cuy. Kalo ini ilang, Go bakal error.
	err = r.DB.Select(&list, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *TransaksiRepository) FindAllWithCount(
	status string,
	limit, offset int,
) ([]domain.Transaksi, int, error) {

	var list []domain.Transaksi
	var total int

	baseQuery := `FROM transaksi WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		baseQuery += " AND status = $" + strconv.Itoa(argIndex)
		args = append(args, status)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
	SELECT id, user_id, apotek_id, total_harga, status, snap_token, payment_url, created_at, updated_at
	` + baseQuery + `
	ORDER BY created_at DESC
	LIMIT $` + strconv.Itoa(argIndex) +
		` OFFSET $` + strconv.Itoa(argIndex+1)

	args = append(args, limit, offset)

	err = r.DB.Select(&list, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *TransaksiRepository) FindByApotekWithCount(
	apotekID uuid.UUID,
	status string,
	limit, offset int,
) ([]domain.Transaksi, int, error) {

	var list []domain.Transaksi
	var total int

	baseQuery := `
	FROM transaksi
	WHERE apotek_id = $1
	`

	args := []interface{}{apotekID}
	argIndex := 2

	if status != "" {
		baseQuery += " AND status = $" + strconv.Itoa(argIndex)
		args = append(args, status)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
	SELECT id, user_id, apotek_id, total_harga, status, snap_token, payment_url, created_at, updated_at
	` + baseQuery + `
	ORDER BY created_at DESC
	LIMIT $` + strconv.Itoa(argIndex) +
		` OFFSET $` + strconv.Itoa(argIndex+1)

	args = append(args, limit, offset)

	err = r.DB.Select(&list, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func GetPaginationAdvanced(c *gin.Context) (page, limit, offset int) {

	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	offset = (page - 1) * limit
	return
}

func (r *TransaksiRepository) FindDetail(transaksiID string) ([]domain.DetailTransaksi, error) {

	var details []domain.DetailTransaksi

	query := `
	SELECT id, transaksi_id, obat_id, jumlah, harga
	FROM detail_transaksi
	WHERE transaksi_id = $1
	`

	err := r.DB.Select(&details, query, uuid.MustParse(transaksiID))
	return details, err
}
