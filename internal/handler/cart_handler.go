package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type CartHandler struct {
	Service *service.CartService
}

// AddToCart godoc
// @Summary Tambah obat ke keranjang
// @Description Menambahkan obat ke keranjang user. Keranjang hanya boleh dari 1 apotek.
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AddCartRequest true "Data item keranjang"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /cart/items [post]
func (h *CartHandler) AddToCart(c *gin.Context) {

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

// GetCart godoc
// @Summary Lihat keranjang aktif
// @Description Mengambil keranjang aktif milik user
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.CartResponse
// @Failure 400 {object} dto.ErrorResponse
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

// UpdateCartItem godoc
// @Summary Update jumlah item keranjang
// @Description Mengubah jumlah obat dalam keranjang
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Cart Item ID"
// @Param request body dto.UpdateCartRequest true "Jumlah baru"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /cart/items/{id} [put]
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

// DeleteCartItem godoc
// @Summary Hapus item dari keranjang
// @Description Menghapus item obat dari keranjang user
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Param id path int true "Cart Item ID"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.ErrorResponse
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
// @Summary Checkout keranjang
// @Description Membuat transaksi dan menghasilkan token pembayaran
// @Tags Checkout
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.CheckoutResponse
// @Failure 400 {object} dto.ErrorResponse
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
