package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sasaefulanwar/medifinder/internal/config"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeGateway struct {
	Creates     atomic.Int32
	StatusCalls atomic.Int32
	Start       time.Time
	Event       PaymentEvent
	Err         error
	CreateErr   error
	ExpireErr   error
	Started     chan struct{}
	Release     chan struct{}
}

func (g *fakeGateway) Create(id string, amount int64, start time.Time) (PaymentSession, error) {
	g.Start = start
	g.Creates.Add(1)
	if g.Started != nil {
		close(g.Started)
		<-g.Release
	}
	return PaymentSession{"test-token", "https://app.sandbox.midtrans.com/snap/v2/test-token"}, g.CreateErr
}
func (g *fakeGateway) Status(id string) (PaymentEvent, error) {
	g.StatusCalls.Add(1)
	if g.Event.Status == "" && g.Err == nil {
		return PaymentEvent{}, ErrPaymentNotFound
	}
	event := g.Event
	event.OrderID = id
	return event, g.Err
}
func (g *fakeGateway) Expire(id string) error {
	if g.ExpireErr != nil {
		return g.ExpireErr
	}
	g.Event.Status = "expire"
	return nil
}

type paymentFixture struct {
	DB                                     *sqlx.DB
	Cart                                   *CartService
	Payment                                *PaymentService
	Gateway                                *fakeGateway
	User, Other, Admin, Pharmacy, Medicine string
}

func paymentDB(t *testing.T) *paymentFixture {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; requires disposable PostgreSQL")
	}
	if _, err := url.Parse(dsn); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join("..", ".."))
	admin, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "payment_test_" + uuid.New().String()
	schema = stringReplaceHyphens(schema)
	if _, err = admin.Exec(`CREATE SCHEMA ` + pq.QuoteIdentifier(schema)); err != nil {
		t.Fatal(err)
	}
	params, err := pq.ParseURL(dsn)
	if err != nil {
		t.Fatal(err)
	}
	db, err := sqlx.Connect("postgres", params+" search_path="+schema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		_, _ = admin.Exec(`DROP SCHEMA ` + pq.QuoteIdentifier(schema) + ` CASCADE`)
		admin.Close()
	})
	db.SetMaxOpenConns(16)
	_, err = db.Exec(`CREATE TABLE users(id uuid PRIMARY KEY,status text,name text,email text,deleted_at timestamptz);
 CREATE TABLE apotek(id uuid PRIMARY KEY,admin_id uuid,verification_status text,nama text,alamat text);
 CREATE TABLE cart(id uuid PRIMARY KEY,user_id uuid,apotek_id uuid);
 CREATE TABLE obat(id uuid PRIMARY KEY,apotek_id uuid,stok int CHECK(stok>=0),reserved_stock int CHECK(reserved_stock>=0),harga numeric,nama text);
 CREATE TABLE cart_items(id uuid PRIMARY KEY,cart_id uuid,obat_id uuid,jumlah int CHECK(jumlah>0));
 CREATE TABLE transaksi(id uuid PRIMARY KEY,user_id uuid,apotek_id uuid,total_harga numeric,status varchar(32),snap_token text,payment_url text,created_at timestamp DEFAULT NOW(),updated_at timestamp DEFAULT NOW());
 CREATE TABLE detail_transaksi(id uuid PRIMARY KEY,transaksi_id uuid,obat_id uuid,jumlah int,harga numeric);`)
	if err != nil {
		t.Fatal(err)
	}
	if err = config.MigratePayments(db); err != nil {
		t.Fatal(err)
	}
	if err = config.MigratePayments(db); err != nil {
		t.Fatal("migration must be repeatable", err)
	}
	f := &paymentFixture{DB: db, Gateway: &fakeGateway{}, User: uuid.NewString(), Other: uuid.NewString(), Admin: uuid.NewString(), Pharmacy: uuid.NewString(), Medicine: uuid.NewString()}
	f.Cart = &CartService{CartRepo: &repository.CartRepository{DB: db}, ObatRepo: &repository.ObatRepository{DB: db}, Gateway: f.Gateway}
	f.Payment = &PaymentService{DB: db, Gateway: f.Gateway}
	if _, err = db.Exec(`INSERT INTO users(id,status)VALUES($1,'active')`, f.Admin); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO apotek(id,admin_id,verification_status,nama,alamat)VALUES($1,$2,'APPROVED','Apotek Uji','Jl. Sehat 10')`, f.Pharmacy, f.Admin); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO obat(id,apotek_id,stok,reserved_stock,harga,nama) VALUES($1,$2,20,0,10000,'Paracetamol 500 mg')`, f.Medicine, f.Pharmacy); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO users(id,status,name,email) VALUES($1,'active','Pelanggan Uji','customer@example.invalid')`, f.User); err != nil {
		t.Fatal(err)
	}
	if err = f.Cart.AddToCart(f.User, f.Medicine, 2); err != nil {
		t.Fatal(err)
	}
	return f
}
func stringReplaceHyphens(s string) string {
	out := []byte(s)
	for i, c := range out {
		if c == '-' {
			out[i] = '_'
		}
	}
	return string(out)
}
func (f *paymentFixture) order(t *testing.T) string {
	t.Helper()
	id, _, _, err := f.Cart.Checkout(f.User)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
func (f *paymentFixture) assertStock(t *testing.T, stock, reserved int) {
	t.Helper()
	var value struct {
		Stock    int `db:"stok"`
		Reserved int `db:"reserved_stock"`
	}
	if err := f.DB.Get(&value, `SELECT stok,reserved_stock FROM obat WHERE id=$1`, f.Medicine); err != nil {
		t.Fatal(err)
	}
	if value.Stock != stock || value.Reserved != reserved {
		t.Fatalf("stock %+v, want %d/%d", value, stock, reserved)
	}
}
func TestCheckoutConcurrentRetryAndCartChanges(t *testing.T) {
	f := paymentDB(t)
	f.Gateway.Started = make(chan struct{})
	f.Gateway.Release = make(chan struct{})
	done := make(chan error, 1)
	var id string
	go func() { var err error; id, _, _, err = f.Cart.Checkout(f.User); done <- err }()
	<-f.Gateway.Started
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _, _, _ = f.Cart.Checkout(f.User) }()
	}
	wg.Wait()
	if err := f.Cart.AddToCart(f.User, f.Medicine, 1); err != nil {
		t.Fatal(err)
	}
	close(f.Gateway.Release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	again, _, _, err := f.Cart.Checkout(f.User)
	if err != nil || again != id {
		t.Fatal("retry must reuse order", err)
	}
	if f.Gateway.Creates.Load() != 1 {
		t.Fatal("duplicate gateway charge")
	}
	var count, quantity int
	_ = f.DB.Get(&count, `SELECT count(*) FROM transaksi`)
	_ = f.DB.Get(&quantity, `SELECT jumlah FROM cart_items`)
	if count != 1 || quantity != 1 {
		t.Fatalf("orders=%d remaining cart=%d", count, quantity)
	}
	f.assertStock(t, 20, 2)
	if _, _, _, err = f.Cart.ContinuePayment(f.Other, id); err == nil {
		t.Fatal("other user accessed payment")
	}
}
func TestCheckoutFailuresKeepCartAndRecoverReservations(t *testing.T) {
	for _, definite := range []bool{true, false} {
		t.Run(map[bool]string{true: "rejected", false: "timeout"}[definite], func(t *testing.T) {
			f := paymentDB(t)
			if definite {
				f.Gateway.CreateErr = gatewayError{401}
			} else {
				f.Gateway.CreateErr = errors.New("timeout")
			}
			id, _, _, err := f.Cart.Checkout(f.User)
			if err == nil || id == "" {
				t.Fatal("expected recoverable order")
			}
			var count int
			_ = f.DB.Get(&count, `SELECT count(*) FROM cart_items`)
			if count != 1 {
				t.Fatal("cart lost on failure")
			}
			if definite {
				f.assertStock(t, 20, 0)
			} else {
				f.assertStock(t, 20, 2)
				_, _, _, _ = f.Cart.Checkout(f.User)
				if f.Gateway.Creates.Load() != 1 {
					t.Fatal("uncertain charge retried")
				}
			}
		})
	}
}
func TestPaymentConcurrentDuplicatesAndRefunds(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	event := PaymentEvent{OrderID: id, Amount: "20000.00", Currency: "IDR", Status: "settlement"}
	var wg sync.WaitGroup
	fail := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); fail <- f.Payment.Apply(event) }()
	}
	wg.Wait()
	close(fail)
	for err := range fail {
		if err != nil {
			t.Fatal(err)
		}
	}
	f.assertStock(t, 18, 0)
	event.Status = "partial_refund"
	event.RefundAmount = "5000"
	if err := f.Payment.Apply(event); err != nil {
		t.Fatal(err)
	}
	event.RefundAmount = "10000"
	if err := f.Payment.Apply(event); err != nil {
		t.Fatal(err)
	}
	var refund int
	_ = f.DB.Get(&refund, `SELECT refunded_amount FROM transaksi WHERE id=$1`, id)
	if refund != 10000 {
		t.Fatal("refund total not updated", refund)
	}
	event.Status = "refund"
	event.RefundAmount = "20000"
	if err := f.Payment.Apply(event); err != nil {
		t.Fatal(err)
	}
	event.Status = "settlement"
	if err := f.Payment.Apply(event); err != nil {
		t.Fatal(err)
	}
	var status string
	_ = f.DB.Get(&status, `SELECT status FROM transaksi WHERE id=$1`, id)
	if status != "refunded" {
		t.Fatal("stale payment replaced refund")
	}
	f.assertStock(t, 18, 0)
}
func TestPaymentRejectsAmountAndCurrency(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	for _, event := range []PaymentEvent{{OrderID: id, Amount: "1", Status: "settlement"}, {OrderID: id, Amount: "20000", Currency: "USD", Status: "settlement"}} {
		if !errors.Is(f.Payment.Apply(event), ErrPaymentAmount) {
			t.Fatal("amount accepted")
		}
	}
	f.assertStock(t, 20, 2)
}
func TestExpiryReconciliationAndLateSettlement(t *testing.T) {
	for _, state := range []string{"settlement", "pending", "notfound", "unavailable"} {
		t.Run(state, func(t *testing.T) {
			f := paymentDB(t)
			id := f.order(t)
			_, _ = f.DB.Exec(`UPDATE transaksi SET expires_at=NOW()-INTERVAL '3 minutes' WHERE id=$1`, id)
			f.Gateway.Event = PaymentEvent{Status: state, Amount: "20000", Currency: "IDR"}
			if state == "notfound" {
				f.Gateway.Err = ErrPaymentNotFound
			}
			if state == "unavailable" {
				f.Gateway.Err = errors.New("offline")
			}
			err := f.Payment.ReconcileExpired()
			if state == "unavailable" {
				if err == nil {
					t.Fatal("error hidden")
				}
				f.assertStock(t, 20, 2)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if state == "settlement" {
				f.assertStock(t, 18, 0)
			} else {
				f.assertStock(t, 20, 0)
				if err = f.Payment.Apply(PaymentEvent{OrderID: id, Status: "settlement", Amount: "20000"}); err != nil {
					t.Fatal(err)
				}
				f.assertStock(t, 18, 0)
			}
		})
	}
}
func TestFulfillmentAuthorizationAndLateStockReview(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	s := &TransaksiService{DB: f.DB, Repo: &repository.TransaksiRepository{DB: f.DB}}
	if s.CanAccess(id, f.Other, "user") || s.CanAccess(id, f.Other, "admin_apotek") || !s.CanAccess(id, f.User, "user") {
		t.Fatal("ownership broken")
	}
	if err := s.UpdateFulfillment(id, f.Admin, "preparing"); err == nil {
		t.Fatal("unpaid order processed")
	}
	if err := f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "settlement"}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateFulfillment(id, f.Admin, "completed"); err == nil {
		t.Fatal("skipped states")
	}
	for _, state := range []string{"preparing", "ready", "completed"} {
		if err := s.UpdateFulfillment(id, f.Admin, state); err != nil {
			t.Fatal(err)
		}
	}
}
func TestLatePaymentWithoutAvailableStockStillRecordsMoney(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	if err := f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "expire"}); err != nil {
		t.Fatal(err)
	}
	_, _ = f.DB.Exec(`UPDATE obat SET stok=1 WHERE id=$1`, f.Medicine)
	if err := f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "settlement"}); err != nil {
		t.Fatal(err)
	}
	var state string
	if err := f.DB.Get(&state, `SELECT status||':'||fulfillment_status FROM transaksi WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if state != "paid:needs_review" {
		t.Fatal(state)
	}
	f.assertStock(t, 1, 0)
}

func TestStockReviewCanBeResolvedAndDoesNotOverwriteReservations(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	_, err := f.DB.Exec(`ALTER TABLE obat ADD COLUMN deskripsi text,ADD COLUMN image_url text,ADD COLUMN updated_at timestamptz`)
	if err != nil {
		t.Fatal(err)
	}
	medicine := &domain.Obat{ID: uuid.MustParse(f.Medicine), Stok: 1, Harga: 10000}
	if err = f.Cart.ObatRepo.Update(medicine); err == nil {
		t.Fatal("reserved stock overwritten")
	}
	medicine.Stok = 20
	medicine.ReservedStock = 0
	if err = f.Cart.ObatRepo.Update(medicine); err != nil {
		t.Fatal(err)
	}
	f.assertStock(t, 20, 2)
	if err = f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "expire"}); err != nil {
		t.Fatal(err)
	}
	_, _ = f.DB.Exec(`UPDATE obat SET stok=1 WHERE id=$1`, f.Medicine)
	if err = f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "settlement"}); err != nil {
		t.Fatal(err)
	}
	if err = f.Payment.ResolveStock(id, f.Other); err == nil {
		t.Fatal("other admin allocated stock")
	}
	if err = f.Payment.ResolveStock(id, f.Admin); err == nil {
		t.Fatal("insufficient stock accepted")
	}
	_, _ = f.DB.Exec(`UPDATE obat SET stok=5 WHERE id=$1`, f.Medicine)
	if err = f.Payment.ResolveStock(id, f.Admin); err != nil {
		t.Fatal(err)
	}
	if err = f.Payment.ResolveStock(id, f.Admin); err != nil {
		t.Fatal("retry not idempotent", err)
	}
	f.assertStock(t, 3, 0)
}

func TestFailedProviderExpiryKeepsReservation(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	_, _ = f.DB.Exec(`UPDATE transaksi SET expires_at=NOW()-INTERVAL '3 minutes' WHERE id=$1`, id)
	f.Gateway.Event = PaymentEvent{Status: "pending", Amount: "20000", Currency: "IDR"}
	f.Gateway.ExpireErr = errors.New("expire permission missing")
	if err := f.Payment.Reconcile(id); err == nil {
		t.Fatal("provider expiry error swallowed")
	}
	f.assertStock(t, 20, 2)
}
func TestInactivePharmacyCannotResumePayment(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	_, _ = f.DB.Exec(`UPDATE users SET status='suspended' WHERE id=$1`, f.Admin)
	if _, _, _, err := f.Cart.ContinuePayment(f.User, id); err == nil {
		t.Fatal("inactive pharmacy checkout accepted")
	}
}
func TestStockExceptionFullRefundDoesNotDeductMedicine(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	if err := f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "expire"}); err != nil {
		t.Fatal(err)
	}
	_, _ = f.DB.Exec(`UPDATE obat SET stok=1 WHERE id=$1`, f.Medicine)
	if err := f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "settlement"}); err != nil {
		t.Fatal(err)
	}
	if err := f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "refund", RefundAmount: "20000"}); err != nil {
		t.Fatal(err)
	}
	f.assertStock(t, 1, 0)
}

func TestResumeSavedPaymentDoesNotDependOnStatusPermission(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	f.Gateway.Err = gatewayError{Code: 403}
	resumed, token, url, err := f.Cart.ContinuePayment(f.User, id)
	if err != nil || resumed != id || token != "test-token" || url == "" {
		t.Fatalf("saved payment must open despite status permission: %s %s %s %v", resumed, token, url, err)
	}
	if f.Gateway.StatusCalls.Load() != 0 || f.Gateway.Creates.Load() != 1 {
		t.Fatal("resuming must neither query status nor create a duplicate session")
	}
	if _, _, _, err = f.Cart.ContinuePayment(f.Other, id); err == nil {
		t.Fatal("other user accessed payment")
	}
	_, _ = f.DB.Exec(`UPDATE transaksi SET expires_at=NOW()-INTERVAL '1 minute' WHERE id=$1`, id)
	if _, _, _, err = f.Cart.ContinuePayment(f.User, id); err == nil {
		t.Fatal("expired session resumed")
	}
	_, _ = f.DB.Exec(`UPDATE transaksi SET expires_at=NOW()+INTERVAL '10 minutes',status='paid' WHERE id=$1`, id)
	if _, _, _, err = f.Cart.ContinuePayment(f.User, id); err == nil {
		t.Fatal("paid order resumed")
	}
	f.assertStock(t, 20, 2)
}

func TestCheckoutReconcilesPreviousOrderBeforeCreatingNewSession(t *testing.T) {
	for _, previous := range []string{"settlement", "notfound"} {
		t.Run(previous, func(t *testing.T) {
			f := paymentDB(t)
			old := f.order(t)
			if err := f.Cart.AddToCart(f.User, f.Medicine, 1); err != nil {
				t.Fatal(err)
			}
			if previous == "settlement" {
				f.Gateway.Event = PaymentEvent{Status: "settlement", Amount: "20000", Currency: "IDR"}
			} else {
				f.Gateway.Err = ErrPaymentNotFound
				_, _ = f.DB.Exec(`UPDATE transaksi SET expires_at=NOW()-INTERVAL '3 minutes' WHERE id=$1`, old)
			}
			id, _, _, err := f.Cart.Checkout(f.User)
			if err != nil || id == old || id == "" {
				t.Fatalf("new checkout reused terminal session: %s %v", id, err)
			}
			var oldState, newState string
			_ = f.DB.Get(&oldState, `SELECT status FROM transaksi WHERE id=$1`, old)
			_ = f.DB.Get(&newState, `SELECT status FROM transaksi WHERE id=$1`, id)
			wantOld := "paid"
			stock := 18
			if previous == "notfound" {
				wantOld = "cancelled"
				stock = 20
			}
			if oldState != wantOld || newState != "pending" {
				t.Fatalf("old=%s new=%s", oldState, newState)
			}
			if f.Gateway.Creates.Load() != 2 {
				t.Fatal("expected a fresh session")
			}
			f.assertStock(t, stock, 1)
		})
	}
}
func TestSnapExpiryUsesZonedDeadlineInsteadOfNaiveCreatedTimestamp(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	var expires time.Time
	if err := f.DB.Get(&expires, `SELECT expires_at FROM transaksi WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if !f.Gateway.Start.Equal(expires.Add(-15 * time.Minute)) {
		t.Fatalf("wrong Snap expiry start %s; deadline %s", f.Gateway.Start, expires)
	}
}
func TestPendingStatusNeverMarksPaymentAsPaid(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	f.Gateway.Event = PaymentEvent{Status: "pending", Amount: "20000", Currency: "IDR"}
	if err := f.Payment.Reconcile(id); err != nil {
		t.Fatal(err)
	}
	var state string
	_ = f.DB.Get(&state, `SELECT status FROM transaksi WHERE id=$1`, id)
	if state != "pending" {
		t.Fatal("unpaid order became paid", state)
	}
	f.assertStock(t, 20, 2)
}

func TestPaymentReceiptSnapshotAndDuplicateNotifications(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	_, _ = f.DB.Exec(`UPDATE obat SET nama='Nama baru',harga=25000 WHERE id=$1`, f.Medicine)
	event := PaymentEvent{OrderID: id, Amount: "20000", Status: "settlement", Currency: "IDR"}
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- f.Payment.Apply(event) }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := f.DB.Get(&count, `SELECT count(*) FROM payment_email_outbox WHERE order_id=$1`, id); err != nil || count != 1 {
		t.Fatal("duplicate or missing receipt", count, err)
	}
	_, _ = f.DB.Exec(`UPDATE users SET email='changed@example.invalid' WHERE id=$1`, f.User)
	var sends atomic.Int32
	f.Payment.SendReceipt = func(to, subject, body, key string) error {
		sends.Add(1)
		if to != "customer@example.invalid" || key != id || !strings.Contains(subject, id) {
			t.Error("wrong recipient or idempotency key")
		}
		for _, want := range []string{"Pelanggan Uji", "Apotek Uji", "Jl. Sehat 10", "Paracetamol 500 mg", "Rp 10.000", "Rp 20.000", ">2</td>", "Lunas"} {
			if !strings.Contains(body, want) {
				t.Errorf("receipt missing %s", want)
			}
		}
		if strings.Contains(body, "Nama baru") || strings.Contains(body, "Rp 25.000") {
			t.Error("receipt used current medicine details instead of purchased snapshot")
		}
		return nil
	}
	errs = make(chan error, 5)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- f.Payment.SendPendingReceipts(10) }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if sends.Load() != 1 {
		t.Fatal("duplicate email delivery", sends.Load())
	}
	if err := f.Payment.SendPendingReceipts(10); err != nil {
		t.Fatal(err)
	}
	if sends.Load() != 1 {
		t.Fatal("sent receipt resent")
	}
}
func TestFailedEmailRetriesWithoutChangingPaidStatus(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	if err := f.Payment.Apply(PaymentEvent{OrderID: id, Amount: "20000", Status: "settlement"}); err != nil {
		t.Fatal(err)
	}
	var keys []string
	f.Payment.SendReceipt = func(to, subject, body, key string) error {
		keys = append(keys, key)
		return errors.New("Brevo unavailable")
	}
	if err := f.Payment.SendPendingReceipts(10); err == nil {
		t.Fatal("delivery failure hidden")
	}
	var state string
	_ = f.DB.Get(&state, `SELECT status FROM transaksi WHERE id=$1`, id)
	if state != "paid" {
		t.Fatal("email failure changed payment")
	}
	f.assertStock(t, 18, 0)
	if err := f.Payment.SendPendingReceipts(10); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 {
		t.Fatal("retry ignored backoff")
	}
	_, _ = f.DB.Exec(`UPDATE payment_email_outbox SET next_attempt_at=NOW() WHERE order_id=$1`, id)
	f.Payment.SendReceipt = func(to, subject, body, key string) error { keys = append(keys, key); return nil }
	if err := f.Payment.SendPendingReceipts(10); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != keys[1] {
		t.Fatal("retry did not reuse idempotency key")
	}
	var sent bool
	_ = f.DB.Get(&sent, `SELECT sent_at IS NOT NULL FROM payment_email_outbox WHERE order_id=$1`, id)
	if !sent {
		t.Fatal("receipt not acknowledged")
	}
}
func TestUnpaidAndRejectedPaymentDoNotQueueReceipt(t *testing.T) {
	f := paymentDB(t)
	id := f.order(t)
	for _, event := range []PaymentEvent{{OrderID: id, Amount: "20000", Status: "pending"}, {OrderID: id, Amount: "20000", Status: "capture", Fraud: "challenge"}, {OrderID: id, Amount: "1", Status: "settlement"}, {OrderID: id, Amount: "20000", Status: "expire"}} {
		_ = f.Payment.Apply(event)
	}
	var count int
	_ = f.DB.Get(&count, `SELECT count(*) FROM payment_email_outbox`)
	if count != 0 {
		t.Fatal("unpaid receipt queued")
	}
}
