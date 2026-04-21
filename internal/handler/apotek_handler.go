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

// SearchNearby godoc
// @Summary Search nearby pharmacies
// @Description Get pharmacies within a certain radius and check if they are open
// @Tags apotek
// @Accept json
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param radius query number false "Radius in km (default 5)"
// @Success 200 {array} dto.ApotekNearbyResponse
// @Router /apotek [get]
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
	// 1. Ambil Query Parameter
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	radiusStr := c.Query("radius")

	// 2. Validasi: Wajib ada lat & lng
	if latStr == "" || lngStr == "" {
		c.JSON(400, gin.H{"error": "Koordinat lat dan lng harus diisi, cuy"})
		return
	}

	lat, errLat := strconv.ParseFloat(latStr, 64)
	lng, errLng := strconv.ParseFloat(lngStr, 64)
	if errLat != nil || errLng != nil {
		c.JSON(400, gin.H{"error": "Format koordinat kagak bener nih"})
		return
	}

	// 3. Set Default & Limit Radius
	radius := 5.0 // default 5km
	if radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			if r > 50 {
				radius = 50 // Batasin biar DB nggak kerja bakti
			} else {
				radius = r
			}
		}
	}

	// 4. Ambil Parameter Pagination
	page, limit, offset := utils.GetPaginationAdvanced(c)

	// 5. Panggil Service (Gunakan nama service yang benar, misal: h.ApotekService)
	// Pastikan Service.SearchNearby nerima (lat, lng, radius, limit, offset)
	data, total, err := h.Service.SearchNearby(lat, lng, radius, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()}) // Pakai 500 kalau errornya dari internal/DB
		return
	}

	// 6. Hitung Metadata
	totalInt64 := int64(total)
	limitInt64 := int64(limit)

	totalPage := (totalInt64 + limitInt64 - 1) / limitInt64

	// 7. Response
	c.JSON(200, gin.H{
		"message": "Berhasil dapet data apotek sekitar, cuy",
		"data":    data,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": totalPage,
		},
	})
}
