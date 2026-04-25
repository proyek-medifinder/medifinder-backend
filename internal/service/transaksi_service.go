package service

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type TransaksiService struct {
	DB         *sqlx.DB
	Repo       *repository.TransaksiRepository
	ApotekRepo *repository.ApotekRepository
}

func (s *TransaksiService) CancelExpiredTransactions() error {

	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ids []uuid.UUID
	err = tx.Select(&ids, `
		SELECT id
		FROM transaksi
		WHERE status = 'pending'
		AND created_at < NOW() - INTERVAL '15 minutes'
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		return err
	}

	for _, id := range ids {

		// 🔥 RELEASE RESERVED STOCK
		_, err = tx.Exec(`
			UPDATE obat o
			SET reserved_stock = o.reserved_stock - d.jumlah
			FROM detail_transaksi d
			WHERE d.obat_id = o.id
			AND d.transaksi_id = $1
		`, id)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			UPDATE transaksi
			SET status = 'cancelled'
			WHERE id = $1 AND status = 'pending'
		`, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *TransaksiService) GetUserHistory(
	userID string,
	status string,
	limit, offset int,
) ([]domain.Transaksi, int, error) {

	return s.Repo.FindByUser(userID, status, limit, offset)
}

func (s *TransaksiService) GetAdminHistory(
	adminID string,
	status string,
	limit, offset int,
) ([]domain.Transaksi, int, error) {

	apotek, err := s.ApotekRepo.FindByAdmin(uuid.MustParse(adminID))
	if err != nil {
		return nil, 0, err
	}

	return s.Repo.FindByApotekWithCount(apotek.ID, status, limit, offset)
}

func (s *TransaksiService) GetAllHistory(
	status string,
	limit, offset int,
) ([]domain.Transaksi, int, error) {

	return s.Repo.FindAllWithCount(status, limit, offset)
}

func (s *TransaksiService) GetDetail(transaksiID string) ([]domain.DetailTransaksi, error) {
	return s.Repo.FindDetail(transaksiID)
}
