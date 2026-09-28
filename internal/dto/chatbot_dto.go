package dto

type ChatRequest struct {
	Message   string  `json:"message" binding:"required"`
	AgeGroup  string  `json:"age_group"` 
	Age       int     `json:"age"`       
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}


type ChatPharmacyResponse struct {
	ApotekID  string  `json:"apotek_id"`
	Nama      string  `json:"nama"`
	Alamat    string  `json:"alamat"`
	Latitude  float64 `json:"latitude"`  
	Longitude float64 `json:"longitude"`  
	Distance  float64 `json:"distance_km"`
	Harga     float64 `json:"harga"`
	Stok      int     `json:"stok"`
	JamBuka   string  `json:"jam_buka"`
	JamTutup  string  `json:"jam_tutup"`
	IsOpen    bool    `json:"is_open"`
}



type ChatResponse struct {
	Reply        string                 `json:"reply"`
	MedicineName string                 `json:"medicine_name"`
	Availability string                 `json:"availability"` 
	IsNearby     bool                   `json:"is_nearby"`
	Pharmacies   []ChatPharmacyResponse `json:"pharmacies"`
}
