package dto

type AdminRequest struct {
	Nama     string `json:"nama" example:"Admin Apotek"`
	Email    string `json:"email" example:"admin@apotek.com"`
	Password string `json:"password" example:"password123"`
}

type AdminResponse struct {
	ID    uint   `json:"id"`
	Nama  string `json:"nama"`
	Email string `json:"email"`
}
