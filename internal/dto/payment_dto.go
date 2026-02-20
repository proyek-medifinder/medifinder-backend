package dto

type PaymentNotification struct {
	OrderID      string `json:"order_id"`
	Status       string `json:"transaction_status"`
	PaymentType  string `json:"payment_type"`
	SignatureKey string `json:"signature_key"`
}
