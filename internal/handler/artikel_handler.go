package handler

import (
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type ArtikelHandler struct {
	Service *service.ArtikelService
}

// GetArticles godoc
// @Summary      Daftar Artikel Kesehatan
// @Description  Mengambil daftar artikel kesehatan yang berstatus PUBLISHED (bisa dari Super Admin atau NewsAPI)
// @Tags         Artikel
// @Accept       json
// @Produce      json
// @Param        page query int false "Nomor halaman (default: 1)"
// @Param        limit query int false "Jumlah data per halaman (default: 10)"
// @Success      200 {object} map[string]interface{} "Berisi array of artikel"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /artikel [get]
func (h *ArtikelHandler) GetArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, err := h.Service.GetPublishedArticles(page, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal mengambil artikel"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// CreateManual godoc
// @Summary      Buat Artikel Manual
// @Description  Membuat artikel kesehatan baru secara manual. Mendukung upload gambar thumbnail.
// @Tags         SuperAdmin Artikel
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        judul formData string true "Judul Artikel"
// @Param        konten formData string true "Isi konten artikel"
// @Param        kategori formData string true "Kategori artikel"
// @Param        status formData string true "Status (DRAFT / PUBLISHED)"
// @Param        thumbnail formData file false "File gambar thumbnail artikel (opsional)"
// @Success      200 {object} map[string]interface{} "Pesan sukses"
// @Failure      400 {object} map[string]interface{} "Input tidak valid"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /superadmin/artikel [post]
func (h *ArtikelHandler) CreateManual(c *gin.Context) {
	superAdminID := c.GetString("user_id")

	var req dto.CreateArtikelRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "Input tidak valid: " + err.Error()})
		return
	}

	var imageURL string
	file, err := c.FormFile("thumbnail")
	if err == nil {
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := "public/uploads/artikel/" + filename

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(500, gin.H{"error": "Gagal menyimpan gambar thumbnail"})
			return
		}
		imageURL = "/uploads/artikel/" + filename
	}

	err = h.Service.CreateManual(superAdminID, req, imageURL)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Artikel berhasil dibuat"})
}

// Delete godoc
// @Summary      Hapus Artikel
// @Description  Menghapus artikel berdasarkan ID
// @Tags         SuperAdmin Artikel
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "ID Artikel"
// @Success      200 {object} map[string]interface{} "Pesan sukses"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /superadmin/artikel/{id} [delete]
func (h *ArtikelHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.DeleteArtikel(id); err != nil {
		c.JSON(500, gin.H{"error": "Gagal menghapus artikel"})
		return
	}
	c.JSON(200, gin.H{"message": "Artikel berhasil dihapus"})
}

// TriggerFetchNews godoc
// @Summary      Fetch Artikel NewsAPI
// @Description  Menjalankan fungsi background untuk menarik artikel kesehatan terbaru dari NewsAPI
// @Tags         SuperAdmin Artikel
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{} "Pesan sukses trigger berjalan"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /superadmin/artikel/fetch [post]
func (h *ArtikelHandler) TriggerFetchNews(c *gin.Context) {
	go h.Service.FetchHealthNews()
	c.JSON(200, gin.H{"message": "Proses fetch artikel API sedang berjalan di background"})
}
