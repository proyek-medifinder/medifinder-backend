package dto

type CheckoutResponse struct {
	Message     string `json:"message" example:"checkout berhasil"`
	TransaksiID string `json:"transaksi_id"`
	SnapToken   string `json:"snap_token"`
	RedirectURL string `json:"redirect_url"`
}
