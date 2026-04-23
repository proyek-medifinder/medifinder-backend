package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type CartHandler struct {
	Service *service.CartService
}

// AddToCart godoc
// @Summary Tambah obat ke keranjang
// @Tags Keranjang
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AddCartRequest true "Data item keranjang"
// @Success 200 {object} map[string]interface{}
// @Router /cart/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.AddCartRequest
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

// GetCart godoc
// @Summary Lihat isi keranjang
// @Description Mengambil semua daftar obat yang ada di keranjang user saat ini
// @Tags Keranjang
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.CartResponse
// @Failure 401 {object} map[string]interface{} "error: unauthorized"
// @Router /cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID := c.GetString("user_id")

	data, err := h.Service.GetCart(userID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, data)
}

// UpdateItem godoc
// @Summary Update jumlah item di keranjang
// @Tags Keranjang
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Cart Item ID"
// @Param request body dto.UpdateCartRequest true "Data update quantity"
// @Success 200 {object} map[string]interface{}
// @Router /cart/items/{id} [put]
func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID := c.GetString("user_id")
	itemID := c.Param("id")

	var req dto.UpdateCartRequest
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

// DeleteItem godoc
// @Summary Hapus item dari keranjang
// @Description Menghapus satu item obat dari keranjang belanja
// @Tags Keranjang
// @Security BearerAuth
// @Param id path string true "Cart Item ID"
// @Success 200 {object} map[string]interface{} "message: item deleted"
// @Router /cart/items/{id} [delete]
func (h *CartHandler) DeleteItem(c *gin.Context) {
	userID := c.GetString("user_id")
	itemID := c.Param("id")

	err := h.Service.DeleteItem(userID, itemID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "item deleted"})
}

// Checkout godoc
// @Summary Proses Checkout
// @Description Membuat transaksi baru dari item-item yang ada di keranjang
// @Tags Keranjang
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{} "message: checkout success, data: transaksi"
// @Router /cart/checkout [post]
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
