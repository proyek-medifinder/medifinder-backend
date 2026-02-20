package dto

type ObatRequest struct {
	Nama  string  `json:"nama" example:"Paracetamol"`
	Stok  int     `json:"stok" example:"50"`
	Harga float64 `json:"harga" example:"5000"`
}

type ObatResponse struct {
	ID       uint    `json:"id"`
	Nama     string  `json:"nama"`
	Stok     int     `json:"stok"`
	Harga    float64 `json:"harga"`
	ApotekID uint    `json:"apotek_id"`
}
