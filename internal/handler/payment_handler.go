package handler

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"log"
	"net/http"
	"os"
)

type PaymentHandler struct{ Service *service.PaymentService }
type MidtransNotification struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
}

func verifyMidtransSignature(req MidtransNotification, secret string) bool {
	if secret == "" {
		return false
	}
	hash := sha512.Sum512([]byte(req.OrderID + req.StatusCode + req.GrossAmount + secret))
	actual, err := hex.DecodeString(req.SignatureKey)
	return err == nil && subtle.ConstantTimeCompare(hash[:], actual) == 1
}
func (h *PaymentHandler) Notification(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
	var req MidtransNotification
	if c.ShouldBindJSON(&req) != nil || req.StatusCode == "" || req.GrossAmount == "" || req.TransactionStatus == "" {
		c.JSON(400, gin.H{"error": "payload tidak valid"})
		return
	}
	if _, err := uuid.Parse(req.OrderID); err != nil {
		c.JSON(400, gin.H{"error": "order_id tidak valid"})
		return
	}
	secret := os.Getenv("MIDTRANS_SIGNATURE_KEY")
	if secret == "" {
		secret = os.Getenv("MIDTRANS_SERVER_KEY")
	}
	if !verifyMidtransSignature(req, secret) {
		c.JSON(401, gin.H{"error": "signature tidak valid"})
		return
	}
	// The legacy signature does not cover transaction_status. Fetch the current
	// authenticated provider state to handle spoofed or out-of-order statuses.
	if err := h.Service.Reconcile(req.OrderID); err != nil {
		log.Printf("Payment notification %s: %v", req.OrderID, err)
		code := 503
		if errors.Is(err, service.ErrPaymentAmount) {
			code = 409
		}
		c.JSON(code, gin.H{"error": "status pembayaran belum dapat dikonfirmasi"})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}
