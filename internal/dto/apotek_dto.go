package dto

type ApotekRequest struct {
	Nama      string  `json:"nama" example:"Apotek Sehat"`
	Alamat    string  `json:"alamat" example:"Jl. Merdeka No 10"`
	Latitude  float64 `json:"latitude" example:"-6.2000"`
	Longitude float64 `json:"longitude" example:"106.8166"`
}

type ApotekResponse struct {
	ID        uint    `json:"id" example:"1"`
	Nama      string  `json:"nama"`
	Alamat    string  `json:"alamat"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Jarak     float64 `json:"jarak_km,omitempty" example:"1.2"`
}
