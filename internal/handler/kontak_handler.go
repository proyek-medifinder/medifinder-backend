package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type KontakHandler struct {
	Service *service.KontakService
}

// SubmitMessage godoc
// @Summary      Kirim Pesan Kontak
// @Description  Memungkinkan user atau guest untuk mengirimkan pesan/pertanyaan ke sistem
// @Tags         Kontak
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateKontakRequest true "Data Pesan"
// @Success      200 {object} map[string]interface{} "Pesan sukses"
// @Failure      400 {object} map[string]interface{} "Input tidak valid"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /kontak [post]
func (h *KontakHandler) SubmitMessage(c *gin.Context) {
	var req dto.CreateKontakRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Format input tidak valid"})
		return
	}

	if err := h.Service.SubmitMessage(req); err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengirim pesan"})
		return
	}

	c.JSON(200, gin.H{"message": "Pesan Anda berhasil dikirim! Tim kami akan segera menghubungi Anda."})
}

// GetMessages godoc
// @Summary      Lihat Semua Pesan Kontak
// @Description  Mengambil daftar pesan masuk dari user (Khusus Super Admin)
// @Tags         SuperAdmin Kontak
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Nomor halaman"
// @Param        limit query int false "Jumlah data per halaman"
// @Success      200 {object} map[string]interface{} "Berisi daftar pesan"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /superadmin/kontak [get]
func (h *KontakHandler) GetMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.Service.GetMessages(page, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengambil daftar pesan"})
		return
	}

	c.JSON(200, gin.H{"data": data})
}

// UpdateStatus godoc
// @Summary      Update Status Pesan
// @Description  Mengubah status pesan (misal: UNREAD jadi RESOLVED)
// @Tags         SuperAdmin Kontak
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "ID Pesan Kontak"
// @Param        request body dto.UpdateKontakStatusRequest true "Status Baru"
// @Success      200 {object} map[string]interface{} "Pesan sukses"
// @Failure      400 {object} map[string]interface{} "Input tidak valid"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /superadmin/kontak/{id} [put]
func (h *KontakHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateKontakStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Status harus berupa UNREAD, READ, atau RESOLVED"})
		return
	}

	if err := h.Service.UpdateStatus(id, req.Status); err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengupdate status pesan"})
		return
	}

	c.JSON(200, gin.H{"message": "Status pesan berhasil diperbarui"})
}
