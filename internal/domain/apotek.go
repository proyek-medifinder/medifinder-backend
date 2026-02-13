package domain

import "github.com/google/uuid"

type Apotek struct {
	ID        uuid.UUID `db:"id"`
	AdminID   uuid.UUID `db:"admin_id"`
	Nama      string    `db:"nama"`
	Alamat    string    `db:"alamat"`
	Latitude  float64   `db:"latitude"`
	Longitude float64   `db:"longitude"`
	Distance  float64   `db:"distance" json:"distance"`
}
