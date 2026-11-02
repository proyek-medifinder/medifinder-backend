package service

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"math"
	"strconv"
	"time"
)

func fmtAmount(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }

func (s *CartService) ContinuePayment(userID, id string) (string, string, string, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", "", "", errors.New("pesanan tidak ditemukan")
	}
	// Reopening a saved Snap session does not require GET Status permission.
	// continueCheckout still checks ownership, local status, expiry and pharmacy
	// availability. Only authenticated reconciliation may mark an order paid.
	return s.continueCheckout(userID, id)
}

// One unfinished checkout per user. A transaction-level advisory lock also
// serializes simultaneous requests arriving at separate backend instances.
func (s *CartService) Checkout(userID string) (string, string, string, error) {
	if s.Gateway == nil && NewMidtransGateway().Key == "" {
		return "", "", "", errors.New("layanan pembayaran belum dikonfigurasi")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", "", "", errors.New("user tidak valid")
	}
	tx, err := s.CartRepo.DB.Beginx()
	if err != nil {
		return "", "", "", err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, userID); err != nil {
		return "", "", "", err
	}
	var pendingID string
	err = tx.Get(&pendingID, `SELECT id FROM transaksi WHERE user_id=$1 AND status='pending' ORDER BY created_at DESC LIMIT 1`, userID)
	if err == nil {
		if err = tx.Commit(); err != nil {
			return "", "", "", err
		}
		// A new checkout must first settle the previous local pending order.
		// Otherwise an already-paid Snap page can masquerade as a new purchase.
		payment := &PaymentService{DB: s.CartRepo.DB, Gateway: s.Gateway}
		if err = payment.Reconcile(pendingID); err != nil && !errors.Is(err, ErrPaymentNotFound) {
			return pendingID, "", "", err
		}
		var status string
		if err = s.CartRepo.DB.Get(&status, `SELECT status FROM transaksi WHERE id=$1 AND user_id=$2`, pendingID, userID); err != nil {
			return pendingID, "", "", err
		}
		if status != "pending" {
			return s.Checkout(userID)
		}
		return s.ContinuePayment(userID, pendingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", "", err
	}
	var cart domain.Cart
	if err = tx.Get(&cart, `SELECT id,user_id,apotek_id FROM cart WHERE user_id=$1 FOR UPDATE`, userID); err != nil {
		return "", "", "", errors.New("keranjang kosong")
	}
	var status string
	err = tx.Get(&status, `SELECT u.status FROM users u JOIN apotek a ON a.admin_id=u.id WHERE a.id=$1 AND u.deleted_at IS NULL AND UPPER(a.verification_status)='APPROVED' FOR SHARE OF u,a`, cart.ApotekID)
	if err != nil || !domain.IsActiveAccount(status) {
		return "", "", "", errors.New("apotek sedang tidak menerima pesanan")
	}
	items, err := s.CartRepo.GetCartWithItemsTx(tx, cart.ID)
	if err != nil {
		return "", "", "", err
	}
	if len(items) == 0 {
		return "", "", "", errors.New("keranjang kosong")
	}
	id := uuid.NewString()
	total := int64(0)
	start := time.Now().Truncate(time.Second)
	type detail struct {
		ID       uuid.UUID
		Name     string
		Quantity int
		Price    int64
	}
	details := []detail{}
	for _, item := range items {
		var obat domain.Obat
		if err = tx.Get(&obat, `SELECT id,apotek_id,nama,stok,reserved_stock,harga FROM obat WHERE id=$1 FOR UPDATE`, item.ObatID); err != nil {
			return "", "", "", err
		}
		if obat.ApotekID != cart.ApotekID || item.Jumlah <= 0 || item.Jumlah > obat.Stok-obat.ReservedStock {
			return "", "", "", errors.New("stok obat tidak cukup atau data keranjang tidak sesuai")
		}
		if math.IsNaN(obat.Harga) || math.IsInf(obat.Harga, 0) || obat.Harga <= 0 || math.Trunc(obat.Harga) != obat.Harga || obat.Harga > 1e12 {
			return "", "", "", errors.New("harga obat harus berupa rupiah bulat yang valid")
		}
		price := int64(obat.Harga)
		if int64(item.Jumlah) > (1e12-total)/price {
			return "", "", "", errors.New("total pesanan terlalu besar")
		}
		total += price * int64(item.Jumlah)
		details = append(details, detail{ID: item.ObatID, Name: obat.Nama, Quantity: item.Jumlah, Price: price})
		if _, err = tx.Exec(`UPDATE obat SET reserved_stock=reserved_stock+$2 WHERE id=$1`, item.ObatID, item.Jumlah); err != nil {
			return "", "", "", err
		}
	}
	_, err = tx.Exec(`INSERT INTO transaksi(id,user_id,apotek_id,total_harga,status,expires_at,reservation_state,checkout_cart_id,cart_cleared,created_at,updated_at) VALUES($1,$2,$3,$4,'pending',$5,'reserved',$6,false,$7,$7)`, id, userUUID, cart.ApotekID, total, start.Add(15*time.Minute), cart.ID, start)
	if err != nil {
		return "", "", "", err
	}
	for _, d := range details {
		if _, err = tx.Exec(`INSERT INTO detail_transaksi(id,transaksi_id,obat_id,jumlah,harga,nama_obat) VALUES($1,$2,$3,$4,$5,$6)`, uuid.New(), id, d.ID, d.Quantity, d.Price, d.Name); err != nil {
			return "", "", "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", "", "", err
	}
	return s.continueCheckout(userID, id)
}

func (s *CartService) continueCheckout(userID, id string) (string, string, string, error) {
	tx, err := s.CartRepo.DB.Beginx()
	if err != nil {
		return id, "", "", err
	}
	defer tx.Rollback()
	var order struct {
		Available bool `db:"available"`
		Status    string
		Token     *string    `db:"snap_token"`
		URL       *string    `db:"payment_url"`
		Total     float64    `db:"total_harga"`
		Created   time.Time  `db:"created_at"`
		Expires   time.Time  `db:"expires_at"`
		Attempt   *time.Time `db:"payment_attempted_at"`
	}
	err = tx.Get(&order, `SELECT status,snap_token,payment_url,total_harga,created_at,expires_at,payment_attempted_at,
 EXISTS(SELECT 1 FROM apotek a JOIN users u ON u.id=a.admin_id WHERE a.id=transaksi.apotek_id AND UPPER(a.verification_status)='APPROVED' AND u.deleted_at IS NULL AND LOWER(BTRIM(u.status::text)) IN ('active','approved')) AS available
 FROM transaksi WHERE id=$1 AND user_id=$2 FOR UPDATE`, id, userID)
	if err != nil {
		return id, "", "", errors.New("pesanan tidak ditemukan")
	}
	if order.Status != "pending" {
		return id, "", "", errors.New("pesanan sudah tidak menunggu pembayaran")
	}
	if !order.Available {
		return id, "", "", errors.New("apotek sedang tidak menerima pesanan; periksa status pembayaran di profil")
	}
	if time.Now().After(order.Expires) {
		return id, "", "", errors.New("waktu pembayaran berakhir; perbarui status pesanan di profil")
	}
	if order.Token != nil && order.URL != nil && *order.URL != "" {
		return id, *order.Token, *order.URL, tx.Commit()
	}
	if order.Attempt != nil {
		return id, "", "", errors.New("pembayaran sedang dikonfirmasi; periksa riwayat di profil, jangan membuat pembayaran ganda")
	}
	if _, err = tx.Exec(`UPDATE transaksi SET payment_attempted_at=NOW() WHERE id=$1`, id); err != nil {
		return id, "", "", err
	}
	if err = tx.Commit(); err != nil {
		return id, "", "", err
	}
	gateway := s.Gateway
	if gateway == nil {
		gateway = NewMidtransGateway()
	}
	// expires_at is timestamptz. Legacy created_at columns may be timestamp
	// without time zone and must not shift the Snap start time by seven hours.
	session, err := gateway.Create(id, int64(order.Total), order.Expires.Add(-15*time.Minute))
	if err != nil {
		if definitiveRejection(err) {
			// A definite rejection created no session. Release reservation; keep cart.
			payment := &PaymentService{DB: s.CartRepo.DB, Gateway: gateway}
			if releaseErr := payment.Apply(PaymentEvent{OrderID: id, Amount: fmtAmount(order.Total), Status: "failure", Currency: "IDR"}); releaseErr != nil {
				return id, "", "", releaseErr
			}
		}
		return id, "", "", err
	}
	save, err := s.CartRepo.DB.Beginx()
	if err != nil {
		return id, "", "", err
	}
	defer save.Rollback()
	var state string
	if err = save.Get(&state, `SELECT status FROM transaksi WHERE id=$1 FOR UPDATE`, id); err != nil {
		return id, "", "", err
	}
	if _, err = save.Exec(`UPDATE transaksi SET snap_token=$2,payment_url=$3,updated_at=NOW() WHERE id=$1`, id, session.Token, session.URL); err != nil {
		return id, "", "", err
	}
	if err = clearCheckoutCart(save, id); err != nil {
		return id, "", "", err
	}
	if err = save.Commit(); err != nil {
		return id, "", "", err
	}
	if state != "pending" {
		return id, "", "", errors.New("status pembayaran berubah; periksa riwayat pesanan")
	}
	return id, session.Token, session.URL, nil
}
