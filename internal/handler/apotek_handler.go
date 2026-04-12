package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"github.com/sasaefulanwar/medifinder/internal/utils"
)

type ApotekHandler struct {
	Service *service.ApotekService
}

type CreateApotekRequest struct {
	Nama      string  `json:"nama" example:"Apotek Sehat"`
	Alamat    string  `json:"alamat" example:"Jl. Raya No. 123"`
	Latitude  float64 `json:"latitude" example:"-6.200000"`
	Longitude float64 `json:"longitude" example:"106.816666"`
	JamBuka   string  `json:"jam_buka" example:"08:00:00"`
	JamTutup  string  `json:"jam_tutup" example:"22:00:00"`
}

// @Summary Buat Apotek Baru
// @Tags Apotek
// @Produce json
// @Router /admin/apotek [post]
func (h *ApotekHandler) Create(c *gin.Context) {
	adminID := c.GetString("user_id")

	var req struct {
		Nama      string  `json:"nama"`
		Alamat    string  `json:"alamat"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		JamBuka   string  `json:"jam_buka"`
		JamTutup  string  `json:"jam_tutup"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.Create(adminID, req.Nama, req.Alamat, req.Latitude, req.Longitude, req.JamBuka, req.JamTutup)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "apotek created"})
}

// GetMyApotek godoc
// @Summary Melihat Profil Apotek Saya (Admin Only)
// @Description Mengambil data lengkap apotek yang dikelola oleh admin yang sedang login
// @Tags Admin Apotek
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.Apotek
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/apotek [get]
func (h *ApotekHandler) GetMyApotek(c *gin.Context) {
	adminID := c.GetString("user_id")

	apotek, err := h.Service.GetByAdmin(adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": apotek})
}

// @Summary Update Profil Apotek Sendiri
// @Tags Apotek
// @Produce json
// @Router /admin/apotek [put]
func (h *ApotekHandler) UpdateMyApotek(c *gin.Context) {
	adminID := c.GetString("user_id")

	var req struct {
		Nama        string  `json:"nama"`
		Alamat      string  `json:"alamat"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		JamBuka     string  `json:"jam_buka"`
		JamTutup    string  `json:"jam_tutup"`
		PhoneNumber string  `json:"phone_number"` // Field baru
		Deskripsi   string  `json:"deskripsi"`    // Field baru
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "format data tidak valid"})
		return
	}

	// Panggil service dengan parameter lengkap
	err := h.Service.Update(adminID, req.Nama, req.Alamat, req.Latitude, req.Longitude, req.JamBuka, req.JamTutup, req.PhoneNumber, req.Deskripsi)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "profil apotek berhasil diupdate"})
}

// @Summary Mencari Apotek Terdekat & Buka
// @Tags Apotek
// @Produce json
// @Router /apotek/nearby [get]
func (h *ApotekHandler) SearchNearby(c *gin.Context) {

	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "5"), 64)

	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.SearchNearby(lat, lng, radius, limit, offset)
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
