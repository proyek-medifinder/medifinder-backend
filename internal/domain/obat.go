package domain

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Obat struct {
	ID            uuid.UUID `db:"id" json:"id"`
	ApotekID      uuid.UUID `db:"apotek_id" json:"apotek_id"`
	Nama          string    `db:"nama" json:"nama"`
	Deskripsi     *string   `db:"deskripsi" json:"deskripsi"`
	Stok          int       `db:"stok" json:"stok"`
	Harga         float64   `db:"harga" json:"harga"`
	ImageURL      *string   `db:"image_url" json:"image_url"`
	ReservedStock int       `db:"reserved_stock" json:"reserved_stock"`
	Kategori      string    `db:"kategori" json:"kategori,omitempty"`
}

// MarshalJSON ensures both lowercase ("id", "image_url") and legacy PascalCase
// ("ID", "Nama", "Stok", "Harga", "ImageURL") exist in the JSON response.
func (o Obat) MarshalJSON() ([]byte, error) {
	type Alias Obat
	return json.Marshal(&struct {
		Alias
		IDUpper        uuid.UUID `json:"ID"`
		ApotekIDUpper  uuid.UUID `json:"ApotekID"`
		NamaUpper      string    `json:"Nama"`
		DeskripsiUpper *string   `json:"Deskripsi,omitempty"`
		StokUpper      int       `json:"Stok"`
		HargaUpper     float64   `json:"Harga"`
		ImageURLUpper  *string   `json:"ImageURL,omitempty"`
	}{
		Alias:          Alias(o),
		IDUpper:        o.ID,
		ApotekIDUpper:  o.ApotekID,
		NamaUpper:      o.Nama,
		DeskripsiUpper: o.Deskripsi,
		StokUpper:      o.Stok,
		HargaUpper:     o.Harga,
		ImageURLUpper:  o.ImageURL,
	})
}
