package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"github.com/sasaefulanwar/medifinder/internal/utils"
)

type TransaksiHandler struct {
	Service *service.TransaksiService
}

// UserHistory godoc
// @Summary Riwayat transaksi user
// @Description Mengambil daftar transaksi milik user yang sedang login
// @Tags Transaksi
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter status transaksi"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} map[string]interface{} "data + pagination"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Failure 500 {object} map[string]interface{} "internal error"
// @Router /transaksi [get]
func (h *TransaksiHandler) UserHistory(c *gin.Context) {

	userID := c.GetString("user_id")
	status := c.Query("status")

	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetUserHistory(userID, status, limit, offset)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	totalPage := (total + limit - 1) / limit

	c.JSON(200, gin.H{
		"data": data,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": totalPage,
		},
	})
}

// Detail godoc
// @Summary Detail transaksi
// @Tags Transaksi
// @Security BearerAuth
// @Param id path string true "Transaksi ID"
// @Success 200 {object} map[string]interface{}
// @Router /transaksi/{id} [get]
func (h *TransaksiHandler) Detail(c *gin.Context) {

	id := c.Param("id")

	data, err := h.Service.GetDetail(id, c.GetString("user_id"), c.GetString("role"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *TransaksiHandler) Refresh(c *gin.Context) {
	id := c.Param("id")
	if !h.Service.CanAccess(id, c.GetString("user_id"), c.GetString("role")) {
		c.JSON(404, gin.H{"error": "pesanan tidak ditemukan"})
		return
	}
	if err := (&service.PaymentService{DB: h.Service.DB}).Reconcile(id); err != nil {
		if !errors.Is(err, service.ErrPaymentNotFound) {
			log.Printf("Payment refresh %s failed: %v", id, err)
			if service.IsPaymentAccessDenied(err) {
				c.JSON(503, gin.H{"error": "Pengecekan status belum tersedia karena izin key Midtrans belum sesuai. Admin perlu memeriksa Server Key dan izin Core API Read."})
				return
			}
			c.JSON(503, gin.H{"error": "Status pembayaran belum dapat disinkronkan. Coba lagi beberapa saat."})
			return
		}
	}
	c.JSON(200, gin.H{"message": "status diperbarui"})
}
func (h *TransaksiHandler) Fulfillment(c *gin.Context) {
	var req struct {
		Status string `json:"status"`
	}
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "status tidak valid"})
		return
	}
	if err := h.Service.UpdateFulfillment(c.Param("id"), c.GetString("user_id"), req.Status); err != nil {
		c.JSON(409, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "status pesanan diperbarui"})
}

// AdminHistory godoc
// @Summary Riwayat transaksi masuk (Admin Apotek)
// @Tags Admin Transaksi
// @Security BearerAuth
// @Router /admin/transaksi [get]
func (h *TransaksiHandler) AdminHistory(c *gin.Context) {

	adminID := c.GetString("user_id")
	status := c.Query("status")

	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetAdminHistory(adminID, status, limit, offset)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	totalPage := (total + limit - 1) / limit

	c.JSON(200, gin.H{
		"data": data,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": totalPage,
		},
	})
}

// SuperAdminHistory godoc
// @Summary Riwayat transaksi masuk (Admin Apotek)
// @Tags Admin Transaksi
// @Security BearerAuth
// @Router /admin/transaksi [get]
func (h *TransaksiHandler) SuperAdminHistory(c *gin.Context) {

	status := c.Query("status")
	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetAllHistory(status, limit, offset)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	totalPage := (total + limit - 1) / limit

	c.JSON(200, gin.H{
		"data": data,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": totalPage,
		},
	})
}
