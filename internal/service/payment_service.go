package service

import (
	"github.com/jmoiron/sqlx"
)

type PaymentService struct {
	DB *sqlx.DB
}

func (s *PaymentService) HandleNotification(
	orderID string,
	transactionStatus string,
	fraudStatus string,
) error {

	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStatus string
	err = tx.Get(&currentStatus, `
		SELECT status
		FROM transaksi
		WHERE id = $1
	`, orderID)
	if err != nil {
		return err
	}

	// 🔒 Idempotency kuat
	if currentStatus != "pending" {
		return tx.Commit()
	}

	switch transactionStatus {

	case "capture":
		if fraudStatus != "accept" {
			return tx.Commit()
		}
		fallthrough

	case "settlement":
		if err := s.markAsPaid(tx, orderID); err != nil {
			return err
		}

	case "cancel", "deny", "expire", "failure":
		_, err = tx.Exec(`
			UPDATE transaksi
			SET status = 'cancelled'
			WHERE id = $1 AND status = 'pending'
		`, orderID)
		if err != nil {
			return err
		}

	case "refund", "partial_refund":
		_, err = tx.Exec(`
			UPDATE transaksi
			SET status = 'refunded'
			WHERE id = $1
		`, orderID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PaymentService) markAsPaid(tx *sqlx.Tx, orderID string) error {

	_, err := tx.Exec(`
		UPDATE obat o
		SET 
			stok = o.stok - d.jumlah,
			reserved_stock = o.reserved_stock - d.jumlah
		FROM detail_transaksi d
		WHERE d.obat_id = o.id
		AND d.transaksi_id = $1
	`, orderID)
	if err != nil {
		return err
	}

	// update status
	_, err = tx.Exec(`
		UPDATE transaksi
		SET status = 'paid'
		WHERE id = $1 AND status = 'pending'
	`, orderID)

	return err
}
