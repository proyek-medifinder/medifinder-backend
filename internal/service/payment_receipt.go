package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/utils"
)

type ReceiptItem struct {
	Name      string
	Quantity  int
	UnitPrice string
	Subtotal  string
}
type PaymentReceipt struct {
	CustomerName    string
	OrderID         string
	PharmacyName    string
	PharmacyAddress string
	Total           string
	ConfirmedAt     string
	NeedsReview     bool
	Items           []ReceiptItem
}

func receiptRupiah(amount string) string {
	value, ok := new(big.Rat).SetString(amount)
	if !ok {
		return "Rp 0"
	}
	digits := value.FloatString(2)
	parts := strings.SplitN(digits, ".", 2)
	integer := parts[0]
	for i := len(integer) - 3; i > 0; i -= 3 {
		integer = integer[:i] + "." + integer[i:]
	}
	if parts[1] != "00" {
		integer += "," + parts[1]
	}
	return "Rp " + integer
}

// Snapshot the recipient and receipt in the same transaction as the paid state.
// No external HTTP calls or template reads can delay the payment commit.
func (s *PaymentService) enqueueReceipt(tx *sqlx.Tx, id string, needsReview bool) error {
	var order struct {
		Name     string
		Email    string
		Pharmacy string `db:"pharmacy"`
		Address  string `db:"address"`
		Total    string `db:"total"`
	}
	err := tx.Get(&order, `SELECT COALESCE(u.name,'Pelanggan') AS name,u.email,
 COALESCE(a.nama,'Apotek') AS pharmacy,COALESCE(a.alamat,'') AS address,t.total_harga::text AS total
 FROM transaksi t JOIN users u ON u.id=t.user_id LEFT JOIN apotek a ON a.id=t.apotek_id WHERE t.id=$1`, id)
	if err != nil {
		return fmt.Errorf("prepare payment receipt: %w", err)
	}
	var items []struct {
		Name     string
		Quantity int    `db:"jumlah"`
		Price    string `db:"price"`
		Subtotal string `db:"subtotal"`
	}
	if err = tx.Select(&items, `SELECT COALESCE(NULLIF(d.nama_obat,''),o.nama,'Obat') AS name,d.jumlah,d.harga::text AS price,(d.harga*d.jumlah)::text AS subtotal
 FROM detail_transaksi d LEFT JOIN obat o ON o.id=d.obat_id WHERE d.transaksi_id=$1 ORDER BY d.id`, id); err != nil {
		return err
	}
	receipt := PaymentReceipt{CustomerName: order.Name, OrderID: id, PharmacyName: order.Pharmacy, PharmacyAddress: order.Address,
		Total: receiptRupiah(order.Total), ConfirmedAt: time.Now().In(time.FixedZone("WIB", 7*3600)).Format("02/01/2006 15:04 WIB"), NeedsReview: needsReview, Items: []ReceiptItem{}}
	for _, item := range items {
		receipt.Items = append(receipt.Items, ReceiptItem{Name: item.Name, Quantity: item.Quantity, UnitPrice: receiptRupiah(item.Price), Subtotal: receiptRupiah(item.Subtotal)})
	}
	payload, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO payment_email_outbox(order_id,recipient,payload) VALUES($1,$2,$3::jsonb) ON CONFLICT(order_id) DO NOTHING`, id, order.Email, string(payload))
	return err
}

// Claim committed messages with a lease. Separate workers cannot send the same
// row concurrently, and restarts recover unfinished messages after the lease.
func (s *PaymentService) SendPendingReceipts(limit int) error {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	send := s.SendReceipt
	if send == nil {
		send = utils.SendEmailNowIdempotent
	}
	var failures []error
	for i := 0; i < limit; i++ {
		var message struct {
			ID      string `db:"order_id"`
			To      string `db:"recipient"`
			Payload []byte `db:"payload"`
			Attempt int    `db:"attempts"`
		}
		err := s.DB.Get(&message, `UPDATE payment_email_outbox SET locked_until=NOW()+INTERVAL '1 minute',attempts=attempts+1
  WHERE order_id=(SELECT order_id FROM payment_email_outbox WHERE sent_at IS NULL AND next_attempt_at<=NOW()
   AND (locked_until IS NULL OR locked_until<NOW()) ORDER BY next_attempt_at,order_id FOR UPDATE SKIP LOCKED LIMIT 1)
  RETURNING order_id,recipient,payload,attempts`)
		if errors.Is(err, sql.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		var receipt PaymentReceipt
		err = json.Unmarshal(message.Payload, &receipt)
		var body string
		if err == nil {
			body, err = utils.ParseTemplate("templates/emails/payment_success.html", receipt)
		}
		if err == nil {
			err = send(message.To, "Pembayaran diterima - Pesanan "+message.ID+" | MediFinder", body, message.ID)
		}
		if err == nil {
			_, err = s.DB.Exec(`UPDATE payment_email_outbox SET sent_at=NOW(),locked_until=NULL,last_error=NULL WHERE order_id=$1 AND attempts=$2 AND sent_at IS NULL`, message.ID, message.Attempt)
			if err != nil {
				return err
			}
		} else {
			delay := 60 << min(message.Attempt-1, 6)
			if delay > 3600 {
				delay = 3600
			}
			detail := err.Error()
			if len(detail) > 500 {
				detail = detail[:500]
			}
			_, saveErr := s.DB.Exec(`UPDATE payment_email_outbox SET locked_until=NULL,next_attempt_at=NOW()+$3::int*INTERVAL '1 second',last_error=$4
   WHERE order_id=$1 AND attempts=$2 AND sent_at IS NULL`, message.ID, message.Attempt, delay, detail)
			if saveErr != nil {
				return saveErr
			}
			failures = append(failures, fmt.Errorf("receipt %s: %w", message.ID, err))
		}
	}
	return errors.Join(failures...)
}
