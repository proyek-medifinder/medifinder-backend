package dto

import "github.com/google/uuid"

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ApotekRequest struct {
	Nama      string  `json:"nama" example:"Apotek Sehat"`
	Alamat    string  `json:"alamat" example:"Jl. Merdeka No 10"`
	Latitude  float64 `json:"latitude" example:"-6.2000"`
	Longitude float64 `json:"longitude" example:"106.8166"`
}

type ApotekDetailResponse struct {
	ID          uuid.UUID      `json:"id"`
	Nama        string         `json:"nama"`
	Alamat      string         `json:"alamat"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	JamBuka     *string        `json:"jam_buka"`
	JamTutup    *string        `json:"jam_tutup"`
	Deskripsi   string         `json:"deskripsi"`
	PhoneNumber *string        `json:"phone_number"`
	PhotoURL    *string        `json:"photo_url"`
	Obats       []ObatResponse `json:"obat"`
}

type CreateApotekRequest struct {
	Nama      string  `json:"nama" binding:"required"`
	Alamat    string  `json:"alamat" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	JamBuka   string  `json:"jam_buka" binding:"required"`
	JamTutup  string  `json:"jam_tutup" binding:"required"`
	PhotoURL  *string `json:"photo_url"`
}

type UpdateApotekRequest struct {
	Nama        string  `json:"nama" example:"Apotek Sehat Sejahtera"`
	Alamat      string  `json:"alamat" example:"Jl. Merdeka No. 45"`
	Latitude    float64 `json:"latitude" example:"-6.200000"`
	Longitude   float64 `json:"longitude" example:"106.816666"`
	JamBuka     *string `json:"jam_buka" example:"08:00:00"`
	JamTutup    *string `json:"jam_tutup" example:"22:00:00"`
	PhoneNumber *string `json:"phone_number" example:"081234567890"`
	Deskripsi   *string `json:"deskripsi" example:"Apotek buka setiap hari."`
	PhotoURL    *string `json:"photo_url" example:"https://example.com/photo.jpg"`
}
