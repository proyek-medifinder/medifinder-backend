package domain

import "github.com/google/uuid"

type Obat struct {
	ID            uuid.UUID `db:"id"`
	ApotekID      uuid.UUID `db:"apotek_id"`
	Nama          string    `db:"nama"`
	Stok          int       `db:"stok"`
	Harga         float64   `db:"harga"`
	ReservedStock int       `db:"reserved_stock"`
	Kategori      string    `db:"kategori"`
}
