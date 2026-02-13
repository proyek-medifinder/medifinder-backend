package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sasaefulanwar/medifinder/internal/domain"
)

type CartRepository struct {
	DB *sqlx.DB
}

type CartItemDetail struct {
	ID     uuid.UUID `db:"id"`
	ObatID uuid.UUID `db:"obat_id"`
	Nama   string    `db:"nama"`
	Harga  int64     `db:"harga"`
	Jumlah int       `db:"jumlah"`
}

func (r *CartRepository) FindByUser(userID string) (*domain.Cart, error) {
	var cart domain.Cart
	query := `SELECT id, user_id, apotek_id FROM cart WHERE user_id=$1`
	err := r.DB.Get(&cart, query, userID)
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) Create(cart *domain.Cart) error {
	query := `
	INSERT INTO cart (id, user_id, apotek_id)
	VALUES (:id, :user_id, :apotek_id)
	`
	_, err := r.DB.NamedExec(query, cart)
	return err
}

func (r *CartRepository) AddItem(item *domain.CartItem) error {
	query := `
	INSERT INTO cart_items (id, cart_id, obat_id, jumlah)
	VALUES (:id, :cart_id, :obat_id, :jumlah)
	`
	_, err := r.DB.NamedExec(query, item)
	return err
}

func (r *CartRepository) FindItem(cartID, obatID string) (*domain.CartItem, error) {
	var item domain.CartItem
	query := `
	SELECT id, cart_id, obat_id, jumlah
	FROM cart_items
	WHERE cart_id=$1 AND obat_id=$2
	`
	err := r.DB.Get(&item, query, cartID, obatID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartRepository) UpdateItem(item *domain.CartItem) error {
	query := `
	UPDATE cart_items
	SET jumlah=:jumlah
	WHERE id=:id
	`
	_, err := r.DB.NamedExec(query, item)
	return err
}

func (r *CartRepository) GetCartItems(cartID string) ([]CartItemDetail, error) {

	var items []CartItemDetail

	query := `
	SELECT 
    ci.id,
    ci.obat_id,
    o.nama,
    o.harga,
    ci.jumlah
FROM cart_items ci
JOIN obat o ON ci.obat_id = o.id
WHERE ci.cart_id = $1
	`

	err := r.DB.Select(&items, query, uuid.MustParse(cartID))
	return items, err
}

func (r *CartRepository) FindItemByID(itemID string) (*domain.CartItem, error) {
	var item domain.CartItem

	query := `
	SELECT id, cart_id, obat_id, jumlah
	FROM cart_items
	WHERE id=$1
	`

	err := r.DB.Get(&item, query, uuid.MustParse(itemID))
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *CartRepository) DeleteItem(itemID string) error {
	query := `DELETE FROM cart_items WHERE id=$1`
	_, err := r.DB.Exec(query, uuid.MustParse(itemID))
	return err
}

func (r *CartRepository) ClearCart(cartID string) error {
	query := `DELETE FROM cart_items WHERE cart_id=$1`
	_, err := r.DB.Exec(query, uuid.MustParse(cartID))
	return err
}

func (r *CartRepository) GetCartWithItemsTx(tx *sqlx.Tx, cartID uuid.UUID) ([]domain.CartItem, error) {

	var items []domain.CartItem

	query := `
	SELECT id, cart_id, obat_id, jumlah
	FROM cart_items
	WHERE cart_id=$1
	`

	err := tx.Select(&items, query, cartID)
	return items, err
}
