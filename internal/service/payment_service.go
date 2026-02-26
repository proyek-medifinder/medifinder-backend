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

	if currentStatus == "paid" {
		return tx.Commit()
	}

	switch transactionStatus {

	case "capture":
		if fraudStatus == "challenge" {
			return tx.Commit()
		}
		fallthrough

	case "settlement":
		if err := s.markAsPaid(tx, orderID); err != nil {
			return err
		}

	case "cancel", "deny", "expire", "failure":
		if currentStatus != "paid" {
			_, err = tx.Exec(`
                UPDATE transaksi
                SET status = 'cancelled'
                WHERE id = $1
            `, orderID)
			if err != nil {
				return err
			}
		}

	case "pending", "authorize":
		// Tidak perlu apa-apa

	case "refund", "partial_refund":
		_, err = tx.Exec(`
            UPDATE transaksi
            SET status = 'refunded'
            WHERE id = $1
        `, orderID)
		if err != nil {
			return err
		}

	default:
	}

	return tx.Commit()
}

func (s *PaymentService) markAsPaid(tx *sqlx.Tx, orderID string) error {

	_, err := tx.Exec(`
        UPDATE transaksi
        SET status = 'paid'
        WHERE id = $1
    `, orderID)

	return err
}
