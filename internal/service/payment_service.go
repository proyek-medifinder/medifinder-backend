package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"math/big"
	"time"
)

var ErrPaymentAmount = errors.New("nominal atau mata uang pembayaran tidak sesuai pesanan")
var ErrPaymentStock = errors.New("stok pembayaran perlu diperiksa")

type PaymentService struct {
	DB          *sqlx.DB
	Gateway     PaymentGateway
	SendReceipt func(to, subject, body, key string) error
}

func (s *PaymentService) gateway() PaymentGateway {
	if s.Gateway != nil {
		return s.Gateway
	}
	return NewMidtransGateway()
}
func sameAmount(a, b string) bool {
	x, ok := new(big.Rat).SetString(a)
	if !ok || x.Sign() <= 0 {
		return false
	}
	y, ok := new(big.Rat).SetString(b)
	return ok && x.Cmp(y) == 0
}
func paymentState(current, status, fraud string) string {
	switch status {
	case "settlement":
		if current == "pending" || current == "cancelled" {
			return "paid"
		}
	case "capture":
		if fraud == "accept" && (current == "pending" || current == "cancelled") {
			return "paid"
		}
	case "cancel", "deny", "expire", "failure":
		if current == "pending" {
			return "cancelled"
		}
	case "refund":
		return "refunded"
	case "partial_refund":
		if current != "refunded" {
			return "partially_refunded"
		}
	}
	return current
}

// Apply accepts only authenticated webhook or GET Status results. Locking the
// order before reading its state prevents duplicate stock changes on retries.
func (s *PaymentService) Apply(event PaymentEvent) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var order struct {
		Status      string
		Amount      string `db:"amount"`
		Reservation string `db:"reservation_state"`
		Refund      string `db:"refunded_amount"`
		Fulfillment string `db:"fulfillment_status"`
	}
	err = tx.Get(&order, `SELECT status,total_harga::text AS amount,reservation_state,refunded_amount::text,fulfillment_status FROM transaksi WHERE id=$1 FOR UPDATE`, event.OrderID)
	if err != nil {
		return err
	}
	if !sameAmount(event.Amount, order.Amount) || (event.Currency != "" && event.Currency != "IDR") {
		return ErrPaymentAmount
	}
	next := paymentState(order.Status, event.Status, event.Fraud)
	if event.Status == "refund" || event.Status == "partial_refund" {
		refund := event.RefundAmount
		if event.Status == "refund" && refund == "" {
			refund = order.Amount
		}
		value, ok := new(big.Rat).SetString(refund)
		total, _ := new(big.Rat).SetString(order.Amount)
		if !ok || value.Sign() <= 0 || value.Cmp(total) > 0 {
			return ErrPaymentAmount
		}
		if event.Status == "refund" && value.Cmp(total) != 0 {
			return ErrPaymentAmount
		}
		if _, err = tx.Exec(`UPDATE transaksi SET refunded_amount=GREATEST(refunded_amount,$2::numeric),updated_at=NOW() WHERE id=$1`, event.OrderID, refund); err != nil {
			return err
		}
	}
	if next == order.Status {
		return tx.Commit()
	}
	reservation := order.Reservation
	needsReview := false
	if next == "paid" || next == "refunded" || next == "partially_refunded" {
		// All stock flows lock the cart before medicines to avoid lock inversion.
		if err = clearCheckoutCart(tx, event.OrderID); err != nil {
			return err
		}
	}
	// A refund is not a physical return. Do not put medicine back into stock.
	if next == "refunded" && order.Fulfillment == "needs_review" {
		// A stock exception never delivered/consumed medicine. A full refund
		// ends that reservation rather than deducting medicine retrospectively.
		if reservation == "reserved" {
			if err = s.adjustStock(tx, event.OrderID, reservation, false); err != nil {
				return err
			}
		}
		reservation = "released"
	} else if next == "paid" || next == "refunded" || next == "partially_refunded" {
		if reservation != "consumed" {
			if err = s.adjustStock(tx, event.OrderID, reservation, true); err != nil {
				if !errors.Is(err, ErrPaymentStock) {
					return err
				}
				needsReview = true
			} else {
				reservation = "consumed"
			}
		}
	} else if next == "cancelled" && reservation == "reserved" {
		if err = s.adjustStock(tx, event.OrderID, reservation, false); err != nil {
			return err
		}
		reservation = "released"
	}
	_, err = tx.Exec(`UPDATE transaksi SET status=$2::text,reservation_state=$3,updated_at=NOW(),
 fulfillment_status=CASE WHEN $2::text='paid' AND fulfillment_status='awaiting_payment' THEN 'paid' ELSE fulfillment_status END WHERE id=$1`, event.OrderID, next, reservation)
	if err != nil {
		return err
	}
	if needsReview {
		if _, err = tx.Exec(`UPDATE transaksi SET fulfillment_status='needs_review' WHERE id=$1`, event.OrderID); err != nil {
			return err
		}
	}
	if next == "paid" {
		if err = s.enqueueReceipt(tx, event.OrderID, needsReview); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// An administrator can retry stock allocation after checking physical stock.
func (s *PaymentService) ResolveStock(id, adminID string) error {
	if _, err := uuid.Parse(id); err != nil {
		return errors.New("pesanan tidak ditemukan")
	}
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var order struct {
		Status      string
		Fulfillment string `db:"fulfillment_status"`
		Reservation string `db:"reservation_state"`
	}
	err = tx.Get(&order, `SELECT t.status,t.fulfillment_status,t.reservation_state FROM transaksi t JOIN apotek a ON a.id=t.apotek_id WHERE t.id=$1 AND a.admin_id=$2 FOR UPDATE OF t`, id, adminID)
	if err != nil {
		return errors.New("pesanan tidak ditemukan")
	}
	if order.Status == "paid" && order.Fulfillment == "paid" && order.Reservation == "consumed" {
		return tx.Commit()
	}
	if order.Status != "paid" || order.Fulfillment != "needs_review" {
		return errors.New("pesanan tidak memerlukan pemeriksaan stok")
	}
	if err = clearCheckoutCart(tx, id); err != nil {
		return err
	}
	if err = s.adjustStock(tx, id, order.Reservation, true); err != nil {
		return errors.New("stok belum cukup; periksa dan perbarui stok obat terlebih dahulu")
	}
	_, err = tx.Exec(`UPDATE transaksi SET reservation_state='consumed',fulfillment_status='paid',updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *PaymentService) adjustStock(tx *sqlx.Tx, id, state string, consume bool) error {
	var items []struct {
		ID       string `db:"obat_id"`
		Quantity int    `db:"quantity"`
	}
	if err := tx.Select(&items, `SELECT obat_id,SUM(jumlah)::int AS quantity FROM detail_transaksi WHERE transaksi_id=$1 GROUP BY obat_id ORDER BY obat_id`, id); err != nil {
		return err
	}
	if len(items) == 0 {
		return errors.New("detail pesanan kosong")
	}
	type change struct {
		ID              string
		Deduct, Release int
	}
	changes := []change{}
	for _, item := range items {
		var stock struct {
			Stock    int `db:"stok"`
			Reserved int `db:"reserved_stock"`
		}
		if err := tx.Get(&stock, `SELECT stok,reserved_stock FROM obat WHERE id=$1 FOR UPDATE`, item.ID); err != nil {
			return err
		}
		release, deduct := 0, 0
		if state == "reserved" {
			release = item.Quantity
		}
		if consume {
			deduct = item.Quantity
		}
		if item.Quantity <= 0 || stock.Reserved < release || (consume && stock.Stock-deduct < stock.Reserved-release) {
			return fmt.Errorf("%w: %s", ErrPaymentStock, id)
		}
		changes = append(changes, change{item.ID, deduct, release})
	}
	for _, c := range changes {
		if _, err := tx.Exec(`UPDATE obat SET stok=stok-$2,reserved_stock=reserved_stock-$3 WHERE id=$1`, c.ID, c.Deduct, c.Release); err != nil {
			return err
		}
	}
	return nil
}

// Clear only purchased quantities, preserving additions made during checkout.
func clearCheckoutCart(tx *sqlx.Tx, id string) error {
	var order struct {
		CartID  *string `db:"checkout_cart_id"`
		Cleared bool    `db:"cart_cleared"`
	}
	if err := tx.Get(&order, `SELECT checkout_cart_id,cart_cleared FROM transaksi WHERE id=$1`, id); err != nil {
		return err
	}
	if order.Cleared || order.CartID == nil {
		return nil
	}
	var cartID string
	if err := tx.Get(&cartID, `SELECT id FROM cart WHERE id=$1 FOR UPDATE`, *order.CartID); err != nil {
		return err
	}
	// Delete zero-result rows first, respecting schemas with CHECK(jumlah>0).
	_, err := tx.Exec(`DELETE FROM cart_items c USING (SELECT obat_id,SUM(jumlah)::int AS quantity FROM detail_transaksi WHERE transaksi_id=$1 GROUP BY obat_id) d WHERE c.cart_id=$2 AND c.obat_id=d.obat_id AND c.jumlah<=d.quantity`, id, cartID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE cart_items c SET jumlah=c.jumlah-d.quantity FROM
 (SELECT obat_id,SUM(jumlah)::int AS quantity FROM detail_transaksi WHERE transaksi_id=$1 GROUP BY obat_id) d
 WHERE c.cart_id=$2 AND c.obat_id=d.obat_id`, id, cartID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE transaksi SET cart_cleared=true WHERE id=$1`, id)
	return err
}
func (s *PaymentService) Reconcile(id string) error {
	event, err := s.gateway().Status(id)
	missing := errors.Is(err, ErrPaymentNotFound)
	if err != nil && !missing {
		return err
	}
	if !missing && event.OrderID != id {
		return errors.New("referensi pembayaran Midtrans tidak cocok")
	}
	if missing || event.Status == "pending" {
		var order struct {
			Status  string
			Amount  string    `db:"amount"`
			Expires time.Time `db:"expires_at"`
		}
		if e := s.DB.Get(&order, `SELECT status,total_harga::text AS amount,expires_at FROM transaksi WHERE id=$1`, id); e != nil {
			return e
		}
		if order.Status == "pending" && time.Now().After(order.Expires.Add(2*time.Minute)) {
			if missing {
				// The unused Snap session is already past its fixed blanket expiry.
				event = PaymentEvent{OrderID: id, Amount: order.Amount, Status: "expire", Currency: "IDR"}
				err = nil
			} else {
				if err = s.gateway().Expire(id); err != nil {
					return err
				}
				event, err = s.gateway().Status(id)
				if err == nil && event.OrderID != id {
					return errors.New("referensi pembayaran Midtrans tidak cocok")
				}
			}
		}
	}
	if err != nil {
		return err
	}
	return s.Apply(event)
}

// Never release reservations on network failure. Workers and manual refresh
// share the same provider reconciliation and atomic stock transition.
func (s *PaymentService) ReconcileExpired() error {
	var ids []string
	if err := s.DB.Select(&ids, `SELECT id FROM transaksi WHERE status='pending' AND expires_at < NOW()-INTERVAL '2 minutes' ORDER BY expires_at LIMIT 50`); err != nil {
		return err
	}
	var failures []error
	for _, id := range ids {
		if err := s.Reconcile(id); err != nil {
			failures = append(failures, fmt.Errorf("rekonsiliasi %s: %w", id, err))
		}
	}
	return errors.Join(failures...)
}
