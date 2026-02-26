package service

import (
	"github.com/google/uuid"
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

	type Detail struct {
		ObatID uuid.UUID `db:"obat_id"`
		Jumlah int       `db:"jumlah"`
	}

	var details []Detail

	err := tx.Select(&details, `
        SELECT obat_id, jumlah
        FROM detail_transaksi
        WHERE transaksi_id = $1
    `, orderID)

	if err != nil {
		return err
	}

	for _, d := range details {
		_, err = tx.Exec(`
            UPDATE obat
            SET stok = stok - $1
            WHERE id = $2
        `, d.Jumlah, d.ObatID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
        UPDATE transaksi
        SET status = 'paid'
        WHERE id = $1
    `, orderID)

	return err
}
