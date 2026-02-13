package domain

import "github.com/google/uuid"

type Cart struct {
	ID       uuid.UUID `db:"id"`
	UserID   uuid.UUID `db:"user_id"`
	ApotekID uuid.UUID `db:"apotek_id"`
}

type CartItem struct {
	ID     uuid.UUID `db:"id"`
	CartID uuid.UUID `db:"cart_id"`
	ObatID uuid.UUID `db:"obat_id"`
	Jumlah int       `db:"jumlah"`
}
