package dto

type AddCartRequest struct {
	ObatID string `json:"obat_id" example:"0d744571-..."`
	Jumlah int    `json:"jumlah" example:"2"`
}

type UpdateCartRequest struct {
	Jumlah int `json:"jumlah" example:"3"`
}

type CartItemResponse struct {
	ID       string  `json:"id"`
	ObatID   string  `json:"obat_id"`
	NamaObat string  `json:"nama_obat"`
	Harga    float64 `json:"harga"`
	Jumlah   int     `json:"jumlah"`
	Subtotal float64 `json:"subtotal"`
}
