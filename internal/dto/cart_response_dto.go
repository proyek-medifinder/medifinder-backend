package dto

type CartResponse struct {
	ApotekID string             `json:"apotek_id"`
	Items    []CartItemResponse `json:"items"`
	Total    float64            `json:"total"`
}
