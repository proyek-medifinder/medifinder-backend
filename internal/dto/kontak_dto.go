package dto

type CreateKontakRequest struct {
	Nama   string `json:"nama" binding:"required"`
	Email  string `json:"email" binding:"required,email"`
	Subjek string `json:"subjek" binding:"required"`
	Pesan  string `json:"pesan" binding:"required"`
}

type UpdateKontakStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=UNREAD READ RESOLVED"`
}
