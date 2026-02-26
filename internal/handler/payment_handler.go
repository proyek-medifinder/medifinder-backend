package handler

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type PaymentHandler struct {
	Service *service.PaymentService
}

type MidtransNotification struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
}

func verifyMidtransSignature(req MidtransNotification, serverKey string) bool {
	raw := req.OrderID + req.StatusCode + req.GrossAmount + serverKey

	hash := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(hash[:])

	return expected == req.SignatureKey
}

// PaymentNotification godoc
// @Summary Callback notifikasi pembayaran
// @Description Endpoint untuk menerima notifikasi pembayaran dari Midtrans
// @Tags Payment
// @Accept json
// @Produce json
// @Param payload body dto.PaymentNotification true "Midtrans notification payload"
// @Success 200 {object} dto.APIResponse
// @Router /payment/notify [post]
func (h *PaymentHandler) Notification(c *gin.Context) {

	var req struct {
		OrderID           string `json:"order_id"`
		StatusCode        string `json:"status_code"`
		GrossAmount       string `json:"gross_amount"`
		SignatureKey      string `json:"signature_key"`
		TransactionStatus string `json:"transaction_status"`
		FraudStatus       string `json:"fraud_status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("❌ invalid JSON payload:", err)
		c.JSON(200, gin.H{"message": "ignored"})
		return
	}

	fmt.Printf("📦 WEBHOOK DARI MIDTRANS: %+v\n", req)

	// validasi field wajib
	if req.OrderID == "" || req.SignatureKey == "" {
		fmt.Println("❌ missing required fields")
		c.JSON(200, gin.H{"message": "ignored"})
		return
	}

	if !verifyMidtransSignature(req, os.Getenv("MIDTRANS_SERVER_KEY")) {
		fmt.Println("❌ invalid signature:", req.OrderID)
		c.JSON(200, gin.H{"message": "invalid signature"})
		return
	}

	if err := h.Service.HandleNotification(
		req.OrderID,
		req.TransactionStatus,
		req.FraudStatus,
	); err != nil {
		fmt.Println("❌ service error:", err)
	}

	c.JSON(200, gin.H{"message": "ok"})
}
