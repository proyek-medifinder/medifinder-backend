package dto

type AddCartRequest struct {
	ObatID uint `json:"obat_id" example:"1"`
	Jumlah int  `json:"jumlah" example:"2"`
}

type UpdateCartRequest struct {
	Jumlah int `json:"jumlah" example:"3"`
}

type CartItemResponse struct {
	ID       uint    `json:"id"`
	ObatID   uint    `json:"obat_id"`
	NamaObat string  `json:"nama_obat"`
	Harga    float64 `json:"harga"`
	Jumlah   int     `json:"jumlah"`
	Subtotal float64 `json:"subtotal"`
}
