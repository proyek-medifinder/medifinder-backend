package domain

import "github.com/google/uuid"

type DetailTransaksi struct {
	ID          uuid.UUID `db:"id" json:"id"`
	TransaksiID uuid.UUID `db:"transaksi_id" json:"transaksi_id"`
	ObatID      uuid.UUID `db:"obat_id" json:"obat_id"`
	NamaObat    string    `db:"nama_obat" json:"nama_obat"`
	Jumlah      int       `db:"jumlah" json:"jumlah"`
	Harga       float64   `db:"harga" json:"harga"`
}
