package service

import (
	"errors"
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
	return (&PaymentService{DB: s.DB}).ReconcileExpired()
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

func (s *TransaksiService) CanAccess(id, userID, role string) bool {
	if _, err := uuid.Parse(id); err != nil {
		return false
	}
	var allowed bool
	err := s.DB.Get(&allowed, `SELECT EXISTS(SELECT 1 FROM transaksi t LEFT JOIN apotek a ON a.id=t.apotek_id
 WHERE t.id=$1 AND (($3='user' AND t.user_id=$2) OR ($3='admin_apotek' AND a.admin_id=$2) OR $3='super_admin'))`, id, userID, role)
	return err == nil && allowed
}
func (s *TransaksiService) GetDetail(transaksiID, userID, role string) ([]domain.DetailTransaksi, error) {
	if !s.CanAccess(transaksiID, userID, role) {
		return nil, errors.New("pesanan tidak ditemukan")
	}
	return s.Repo.FindDetail(transaksiID)
}

func (s *TransaksiService) UpdateFulfillment(id, adminID, next string) error {
	if next=="paid"{return (&PaymentService{DB:s.DB}).ResolveStock(id,adminID)}
	previous := map[string]string{"preparing": "paid", "ready": "preparing", "completed": "ready"}[next]
	if previous == "" {
		return errors.New("status pesanan tidak valid")
	}
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("pesanan tidak ditemukan")
	}
	result, err := s.DB.Exec(`UPDATE transaksi t SET fulfillment_status=$3,updated_at=NOW() FROM apotek a
 WHERE t.id=$1 AND a.id=t.apotek_id AND a.admin_id=$2 AND t.status='paid' AND t.reservation_state='consumed'
 AND (t.fulfillment_status=$4 OR t.fulfillment_status=$3)`, id, adminID, next, previous)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("pesanan belum dibayar atau urutan status tidak sesuai")
	}
	return nil
}
