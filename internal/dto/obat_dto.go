package dto

type ObatRequest struct {
	Nama      string  `json:"nama" form:"nama" example:"Paracetamol"`
	Deskripsi *string `json:"deskripsi" form:"deskripsi" example:"Obat penurun demam"`
	Stok      int     `json:"stok" form:"stok" example:"50"`
	Harga     float64 `json:"harga" form:"harga" example:"5000"`
	ImageURL  *string `json:"image_url" form:"image_url"`
}

type ObatResponse struct {
	ID        string  `json:"id"`
	ApotekID  string  `json:"apotek_id"`
	Nama      string  `json:"nama"`
	Deskripsi *string `json:"deskripsi,omitempty"`
	Stok      int     `json:"stok"`
	Harga     float64 `json:"harga"`
	ImageURL  *string `json:"image_url,omitempty"`
}

