package dto

const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusReady     = "ready"
	StatusCompleted = "completed"
	StatusCanceled  = "canceled"
)

type DetailItemTransaksi struct {
	NamaObat string  `json:"nama_obat"`
	Jumlah   int     `json:"jumlah"`
	Harga    float64 `json:"harga"`
	Subtotal float64 `json:"subtotal"`
}

type TransaksiResponse struct {
	ID      uint                  `json:"id"`
	Apotek  string                `json:"apotek"`
	Total   float64               `json:"total"`
	Status  string                `json:"status"`
	Tanggal string                `json:"tanggal"`
	Items   []DetailItemTransaksi `json:"items,omitempty"`
}
