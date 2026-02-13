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

	var transaksiIDs []uuid.UUID

	err = tx.Select(&transaksiIDs, `
		SELECT id
		FROM transaksi
		WHERE status = 'pending'
		AND expired_at < NOW()
	`)
	if err != nil {
		return err
	}

	for _, transaksiID := range transaksiIDs {

		type Detail struct {
			ObatID uuid.UUID `db:"obat_id"`
			Jumlah int       `db:"jumlah"`
		}

		var details []Detail

		err = tx.Select(&details, `
			SELECT obat_id, jumlah
			FROM detail_transaksi
			WHERE transaksi_id = $1
		`, transaksiID)
		if err != nil {
			return err
		}

		for _, d := range details {
			_, err = tx.Exec(`
				UPDATE obat
				SET stok = stok + $1
				WHERE id = $2
			`, d.Jumlah, d.ObatID)
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(`
			UPDATE transaksi
			SET status = 'cancelled'
			WHERE id = $1
		`, transaksiID)
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

	apotek, err := s.ApotekRepo.FindByAdmin(adminID)
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