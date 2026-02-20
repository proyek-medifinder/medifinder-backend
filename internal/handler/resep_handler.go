package handler

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"github.com/sasaefulanwar/medifinder/internal/utils"
)

type ResepHandler struct {
	Service *service.ResepService
}

// UploadResep godoc
// @Summary Upload resep dokter
// @Description User mengunggah resep sebagai dokumen pendukung transaksi
// @Tags Resep
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File resep (max 5MB)"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /resep [post]
func (h *ResepHandler) Upload(c *gin.Context) {

	userID := c.GetString("user_id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file required"})
		return
	}

	if file.Size > 5*1024*1024 {
		c.JSON(400, gin.H{"error": "max file 5MB"})
		return
	}

	uploadDir := "uploads"
	os.MkdirAll(uploadDir, os.ModePerm)

	filePath := filepath.Join(uploadDir, file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		c.JSON(500, gin.H{"error": "upload failed"})
		return
	}

	err = h.Service.Create(userID, filePath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "resep uploaded"})
}

// ListResep godoc
// @Summary Daftar resep
// @Description Admin apotek melihat daftar resep yang diunggah user
// @Tags Admin Resep
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Limit data (max 50)"
// @Success 200 {object} dto.PaginatedResepResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /admin/resep [get]
func (h *ResepHandler) List(c *gin.Context) {

	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.List(limit, offset)
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

// UpdateStatusResep godoc
// @Summary Update status resep
// @Description Admin apotek memverifikasi atau menolak resep
// @Tags Admin Resep
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID Resep"
// @Param request body dto.UpdateResepStatusRequest true "Status resep"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /admin/resep/{id} [put]
func (h *ResepHandler) UpdateStatus(c *gin.Context) {

	id := c.Param("id")

	var req struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.UpdateStatus(id, req.Status)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "status updated"})
}
