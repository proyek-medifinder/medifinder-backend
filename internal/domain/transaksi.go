package domain

import (
	"time"

	"github.com/google/uuid"
)

type Transaksi struct {
	RefundedAmount    float64   `db:"refunded_amount" json:"refunded_amount"`
	ExpiresAt         time.Time `db:"expires_at" json:"expires_at"`
	FulfillmentStatus string    `db:"fulfillment_status" json:"fulfillment_status"`
	ID                uuid.UUID `db:"id" json:"id"`
	UserID            uuid.UUID `db:"user_id" json:"user_id"`
	ApotekID          uuid.UUID `db:"apotek_id" json:"apotek_id"`
	TotalHarga        float64   `db:"total_harga" json:"total_harga"`
	Status            string    `db:"status" json:"status"`
	SnapToken         *string   `db:"snap_token" json:"snap_token,omitempty"`
	PaymentURL        *string   `db:"payment_url" json:"payment_url,omitempty"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}
