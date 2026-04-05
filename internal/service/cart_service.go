package service

import (
	"errors"
	"os"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type CartService struct {
	CartRepo *repository.CartRepository
	ObatRepo *repository.ObatRepository
}

func (s *CartService) AddToCart(userID, obatID string, jumlah int) error {
	// 1. Validasi userID (Biar gak crash kalau kosong)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("user id tidak valid atau belum login")
	}

	// 2. Validasi obatID
	obatUUID, err := uuid.Parse(obatID)
	if err != nil {
		return errors.New("obat id tidak valid")
	}

	// 3. Cari obat (Ganti pemanggilan ke repo pake obatUUID)
	obat, err := s.ObatRepo.FindByID(obatUUID.String())
	if err != nil {
		return errors.New("obat not found")
	}

	if jumlah > obat.Stok {
		return errors.New("stok tidak cukup")
	}

	// 4. Cari keranjang (Ganti pemanggilan ke repo pake userUUID)
	cart, err := s.CartRepo.FindByUser(userUUID.String())

	if err != nil {
		newCart := &domain.Cart{
			ID:       uuid.New(),
			UserID:   userUUID, // Pake yang udah di-parse tadi
			ApotekID: obat.ApotekID,
		}

		if err := s.CartRepo.Create(newCart); err != nil {
			return err
		}

		cart = newCart
	}

	if cart.ApotekID != obat.ApotekID {
		items, err := s.CartRepo.GetCartItems(cart.ID.String())
		if err == nil && len(items) == 0 {
			if err := s.CartRepo.UpdateApotekID(cart.ID.String(), obat.ApotekID.String()); err != nil {
				return errors.New("gagal mengupdate data keranjang")
			}
			cart.ApotekID = obat.ApotekID
		} else {
			return errors.New("tidak bisa campur apotek berbeda")
		}
	}

	item, err := s.CartRepo.FindItem(cart.ID.String(), obatID)

	if err != nil {
		newItem := &domain.CartItem{
			ID:     uuid.New(),
			CartID: cart.ID,
			ObatID: obat.ID,
			Jumlah: jumlah,
		}

		return s.CartRepo.AddItem(newItem)
	}

	item.Jumlah += jumlah

	if item.Jumlah > obat.Stok {
		return errors.New("stok tidak cukup")
	}

	return s.CartRepo.UpdateItem(item)
}

func (s *CartService) GetCart(userID string) (interface{}, error) {

	cart, err := s.CartRepo.FindByUser(userID)
	if err != nil {
		return nil, errors.New("cart kosong")
	}

	items, err := s.CartRepo.GetCartItems(cart.ID.String())
	if err != nil {
		return nil, err
	}

	var total int64 = 0

	var result []map[string]interface{}

	for _, item := range items {

		subtotal := item.Harga * int64(item.Jumlah)
		total += subtotal

		result = append(result, map[string]interface{}{
			"item_id":  item.ID,
			"obat_id":  item.ObatID,
			"nama":     item.Nama,
			"harga":    item.Harga,
			"jumlah":   item.Jumlah,
			"subtotal": subtotal,
		})

	}

	return map[string]interface{}{
		"items": result,
		"total": total,
	}, nil
}

func (s *CartService) UpdateItem(userID, itemID string, jumlah int) error {

	if jumlah <= 0 {
		return errors.New("jumlah tidak valid")
	}

	cart, err := s.CartRepo.FindByUser(userID)
	if err != nil {
		return errors.New("cart tidak ditemukan")
	}

	item, err := s.CartRepo.FindItemByID(itemID)
	if err != nil {
		return errors.New("item tidak ditemukan")
	}

	if item.CartID != cart.ID {
		return errors.New("forbidden")
	}

	obat, err := s.ObatRepo.FindByID(item.ObatID.String())
	if err != nil {
		return err
	}

	if jumlah > obat.Stok {
		return errors.New("stok tidak cukup")
	}

	item.Jumlah = jumlah

	return s.CartRepo.UpdateItem(item)
}

func (s *CartService) DeleteItem(userID, itemID string) error {

	cart, err := s.CartRepo.FindByUser(userID)
	if err != nil {
		return errors.New("cart tidak ditemukan")
	}

	item, err := s.CartRepo.FindItemByID(itemID)
	if err != nil {
		return errors.New("item tidak ditemukan")
	}

	if item.CartID != cart.ID {
		return errors.New("forbidden")
	}

	return s.CartRepo.DeleteItem(itemID)
}

func (s *CartService) Checkout(userID string) (string, string, string, error) {

	cart, err := s.CartRepo.FindByUser(userID)
	if err != nil {
		return "", "", "", errors.New("cart kosong")
	}

	tx, err := s.CartRepo.DB.Beginx()
	if err != nil {
		return "", "", "", err
	}
	defer tx.Rollback()

	items, err := s.CartRepo.GetCartWithItemsTx(tx, cart.ID)
	if err != nil || len(items) == 0 {
		return "", "", "", errors.New("cart kosong")
	}

	var total int64 = 0
	transaksiID := uuid.New()

	type DetailData struct {
		ObatID uuid.UUID
		Jumlah int
		Harga  int64
	}
	var detailsToInsert []DetailData

	for _, item := range items {
		var obat domain.Obat

		// Gunakan tx.Get dan FOR UPDATE untuk mencegah overselling (Race Condition)
		err = tx.Get(&obat, "SELECT id, apotek_id, nama, stok, harga FROM obat WHERE id=$1 FOR UPDATE", item.ObatID)
		if err != nil {
			return "", "", "", errors.New("gagal mengambil data obat")
		}

		if item.Jumlah > obat.Stok {
			return "", "", "", errors.New("stok tidak cukup untuk obat: " + obat.Nama)
		}

		total += obat.Harga * int64(item.Jumlah)

		// RESERVE STOCK: Langsung kurangi stok di database sekarang juga
		_, err = tx.Exec("UPDATE obat SET stok = stok - $1 WHERE id = $2", item.Jumlah, item.ObatID)
		if err != nil {
			return "", "", "", err
		}

		detailsToInsert = append(detailsToInsert, DetailData{
			ObatID: item.ObatID,
			Jumlah: item.Jumlah,
			Harga:  obat.Harga,
		})
	}

	// 2. Insert tabel transaksi
	// 2. Insert tabel transaksi (HAPUS expired_at karena udah gak ada di DB)
	_, err = tx.Exec(`
	INSERT INTO transaksi (id, user_id, apotek_id, total)
	VALUES ($1, $2, $3, $4)
	`, transaksiID, uuid.MustParse(userID), cart.ApotekID, total)
	if err != nil {
		return "", "", "", err
	}

	// 3. Insert tabel detail_transaksi
	for _, d := range detailsToInsert {
		_, err = tx.Exec(`
		INSERT INTO detail_transaksi (id, transaksi_id, obat_id, jumlah, harga)
		VALUES ($1, $2, $3, $4, $5)
		`, uuid.New(), transaksiID, d.ObatID, d.Jumlah, d.Harga)
		if err != nil {
			return "", "", "", err
		}
	}

	// HAPUS cart_item (TYPO FIX: nama tabel di DB lu itu cart_item, bukan cart_items)
	_, err = tx.Exec(`DELETE FROM cart_item WHERE cart_id=$1`, cart.ID)
	if err != nil {
		return "", "", "", err
	}

	// ================= MIDTRANS SNAP =================
	// Kita pindahin Midtrans ke atas SANGAT PENTING sebelum tx.Commit()
	var snapClient snap.Client
	isProduction := os.Getenv("MIDTRANS_IS_PRODUCTION") == "true"
	if isProduction {
		snapClient.New(os.Getenv("MIDTRANS_SERVER_KEY"), midtrans.Production)
	} else {
		snapClient.New(os.Getenv("MIDTRANS_SERVER_KEY"), midtrans.Sandbox)
	}

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  transaksiID.String(),
			GrossAmt: total,
		},
	}

	snapResp, midtransErr := snapClient.CreateTransaction(req)
	if midtransErr != nil {
		return "", "", "", midtransErr
	}

	// 4. UPDATE transaksi buat nyimpen Token & URL dari Midtrans
	_, err = tx.Exec(`
		UPDATE transaksi
		SET snap_token = $1, payment_url = $2
		WHERE id = $3
	`, snapResp.Token, snapResp.RedirectURL, transaksiID)
	if err != nil {
		return "", "", "", err
	}

	// 5. COMMIT SEMUA TRANSAKSI KE DATABASE
	if err := tx.Commit(); err != nil {
		return "", "", "", err
	}

	return transaksiID.String(), snapResp.Token, snapResp.RedirectURL, nil

}
