package domain

import "github.com/google/uuid"

type DetailTransaksi struct {
	ID          uuid.UUID `db:"id" json:"id"`
	TransaksiID uuid.UUID `db:"transaksi_id" json:"transaksi_id"`
	ObatID      uuid.UUID `db:"obat_id" json:"obat_id"`
	Jumlah      int       `db:"jumlah" json:"jumlah"`
	Harga       int64     `db:"harga" json:"harga"`
}
