package handler

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type PaymentHandler struct {
	Service *service.PaymentService
}

func (h *PaymentHandler) Notification(c *gin.Context) {
	var req struct {
		OrderID           string `json:"order_id"`
		StatusCode        string `json:"status_code"`
		GrossAmount       string `json:"gross_amount"`
		SignatureKey      string `json:"signature_key"`
		TransactionStatus string `json:"transaction_status"`
		FraudStatus       string `json:"fraud_status"`
	}

	body, _ := io.ReadAll(c.Request.Body)
	fmt.Println("==== RAW BODY FROM MIDTRANS ====")
	fmt.Println(string(body))
	fmt.Println("================================")

	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("❌ JSON BIND ERROR:", err)
		c.JSON(200, gin.H{"message": "ignored"})
		return
	}

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")

	fmt.Println("MASUK HANDLER")

	raw := req.OrderID + req.StatusCode + req.GrossAmount + serverKey
	hash := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(hash[:])

	if expected != req.SignatureKey {
		fmt.Println("❌ INVALID SIGNATURE for order:", req.OrderID)
		c.JSON(200, gin.H{"message": "invalid signature"})
		return
	}

	err := h.Service.HandleNotification(
		req.OrderID,
		req.TransactionStatus,
		req.FraudStatus,
	)

	if err != nil {
		fmt.Println("❌ SERVICE ERROR:", err)
	}

	c.JSON(200, gin.H{"message": "ok"})
}
