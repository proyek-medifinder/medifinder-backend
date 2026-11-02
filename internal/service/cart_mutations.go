package service

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

func (s *CartService) cartMutation(userID string, operation func(*sqlx.Tx) error) error {
	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("user tidak valid")
	}
	tx, err := s.CartRepo.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, userID); err != nil {
		return err
	}
	if err = operation(tx); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *CartService) AddToCart(userID, obatID string, jumlah int) error {
	if jumlah <= 0 {
		return errors.New("jumlah harus lebih dari nol")
	}
	if _, err := uuid.Parse(obatID); err != nil {
		return errors.New("obat tidak valid")
	}
	return s.cartMutation(userID, func(tx *sqlx.Tx) error {
		// Read its pharmacy first, but validate it again after locking the medicine.
		var pharmacy string
		if err := tx.Get(&pharmacy, `SELECT apotek_id FROM obat WHERE id=$1`, obatID); err != nil {
			return errors.New("obat tidak ditemukan")
		}
		var cart domain.Cart
		err := tx.Get(&cart, `SELECT id,user_id,apotek_id FROM cart WHERE user_id=$1 FOR UPDATE`, userID)
		if errors.Is(err, sql.ErrNoRows) {
			cart = domain.Cart{ID: uuid.New(), UserID: uuid.MustParse(userID), ApotekID: uuid.MustParse(pharmacy)}
			_, err = tx.Exec(`INSERT INTO cart(id,user_id,apotek_id) VALUES($1,$2,$3)`, cart.ID, cart.UserID, cart.ApotekID)
		}
		if err != nil {
			return err
		}
		if cart.ApotekID.String() != pharmacy {
			var count int
			if err = tx.Get(&count, `SELECT count(*) FROM cart_items WHERE cart_id=$1`, cart.ID); err != nil {
				return err
			}
			if count > 0 {
				return errors.New("tidak bisa campur apotek berbeda")
			}
			if _, err = tx.Exec(`UPDATE cart SET apotek_id=$2 WHERE id=$1`, cart.ID, pharmacy); err != nil {
				return err
			}
		}
		var status string
		if err = tx.Get(&status, `SELECT u.status FROM apotek a JOIN users u ON u.id=a.admin_id WHERE a.id=$1 AND UPPER(a.verification_status)='APPROVED' AND u.deleted_at IS NULL FOR SHARE OF u,a`, pharmacy); err != nil || !domain.IsActiveAccount(status) {
			return errors.New("apotek sedang tidak menerima pesanan")
		}
		var medicine domain.Obat
		if err = tx.Get(&medicine, `SELECT apotek_id,stok,reserved_stock FROM obat WHERE id=$1 FOR UPDATE`, obatID); err != nil {
			return err
		}
		if medicine.ApotekID.String() != pharmacy {
			return errors.New("data obat berubah; muat ulang halaman")
		}
		var item domain.CartItem
		err = tx.Get(&item, `SELECT id,cart_id,obat_id,jumlah FROM cart_items WHERE cart_id=$1 AND obat_id=$2 FOR UPDATE`, cart.ID, obatID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if jumlah > medicine.Stok-medicine.ReservedStock-item.Jumlah {
			return errors.New("stok tidak cukup")
		}
		if errors.Is(err, sql.ErrNoRows) {
			_, err = tx.Exec(`INSERT INTO cart_items(id,cart_id,obat_id,jumlah) VALUES($1,$2,$3,$4)`, uuid.New(), cart.ID, obatID, jumlah)
		} else {
			_, err = tx.Exec(`UPDATE cart_items SET jumlah=jumlah+$2 WHERE id=$1`, item.ID, jumlah)
		}
		return err
	})
}
func (s *CartService) UpdateItem(userID, itemID string, jumlah int) error {
	if jumlah <= 0 {
		return errors.New("jumlah tidak valid")
	}
	return s.changeCartItem(userID, itemID, &jumlah)
}
func (s *CartService) DeleteItem(userID, itemID string) error {
	return s.changeCartItem(userID, itemID, nil)
}
func (s *CartService) changeCartItem(userID, itemID string, quantity *int) error {
	if _, err := uuid.Parse(itemID); err != nil {
		return errors.New("item tidak valid")
	}
	return s.cartMutation(userID, func(tx *sqlx.Tx) error {
		var cartID string
		if err := tx.Get(&cartID, `SELECT id FROM cart WHERE user_id=$1 FOR UPDATE`, userID); err != nil {
			return errors.New("keranjang tidak ditemukan")
		}
		var item domain.CartItem
		if err := tx.Get(&item, `SELECT id,cart_id,obat_id,jumlah FROM cart_items WHERE id=$1 AND cart_id=$2 FOR UPDATE`, itemID, cartID); err != nil {
			return errors.New("item tidak ditemukan")
		}
		if quantity == nil {
			_, err := tx.Exec(`DELETE FROM cart_items WHERE id=$1`, itemID)
			return err
		}
		var available int
		if err := tx.Get(&available, `SELECT stok-reserved_stock FROM obat WHERE id=$1 FOR UPDATE`, item.ObatID); err != nil {
			return err
		}
		if *quantity > available {
			return errors.New("stok tidak cukup")
		}
		_, err := tx.Exec(`UPDATE cart_items SET jumlah=$2 WHERE id=$1`, itemID, *quantity)
		return err
	})
}
