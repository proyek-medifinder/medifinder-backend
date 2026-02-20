package dto

type UpdateResepStatusRequest struct {
	Status string `json:"status" binding:"required" example:"approved"`
}

type ResepResponse struct {
	ID          uint   `json:"id"`
	TransaksiID uint   `json:"transaksi_id"`
	FileURL     string `json:"file_url"`
	Status      string `json:"status"`
}
