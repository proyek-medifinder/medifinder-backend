package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type CartHandler struct {
	Service *service.CartService
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	// Ganti "user_id" jadi "id" biar sesuai middleware
	userID := c.GetString("user_id")

	var req struct {
		ObatID string `json:"obat_id"`
		Jumlah int    `json:"jumlah"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.AddToCart(userID, req.ObatID, req.Jumlah)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "added to cart"})
}

func (h *CartHandler) GetCart(c *gin.Context) {
	// Ganti "user_id" jadi "id"
	userID := c.GetString("user_id")

	data, err := h.Service.GetCart(userID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, data)
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID := c.GetString("user_id")
	itemID := c.Param("id")

	var req struct {
		Jumlah int `json:"jumlah"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.UpdateItem(userID, itemID, req.Jumlah)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "item updated"})
}

func (h *CartHandler) DeleteItem(c *gin.Context) {
	// Ganti "user_id" jadi "id"
	userID := c.GetString("user_id")
	itemID := c.Param("id")

	err := h.Service.DeleteItem(userID, itemID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "item deleted"})
}

func (h *CartHandler) Checkout(c *gin.Context) {
	userID := c.GetString("user_id")

	transaksiID, snapToken, redirectURL, err := h.Service.Checkout(userID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":      "checkout berhasil",
		"transaksi_id": transaksiID,
		"snap_token":   snapToken,
		"redirect_url": redirectURL,
	})
}
