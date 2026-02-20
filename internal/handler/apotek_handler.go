package handler

import (
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
}

// SearchNearby godoc
// @Summary Cari apotek terdekat
// @Tags Apotek
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {array} dto.ApotekResponse
// @Router /apotek [get]
func (h *ApotekHandler) Create(c *gin.Context) {
	adminID := c.GetString("user_id")

	var req struct {
		Nama      string  `json:"nama"`
		Alamat    string  `json:"alamat"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.Create(adminID, req.Nama, req.Alamat, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "apotek created"})
}

// @Summary Melihat Apotek Saya
// @Description Melihat Apotek Berada di Akun Admin yang Sedang Login
// @Tags Apotek
// @Produce json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} domain.Apotek
// @Failure 404 {object} map[string]string "error: not found"
// @Router /apotek/me [get]
func (h *ApotekHandler) GetMyApotek(c *gin.Context) {
	adminID := c.GetString("user_id")

	apotek, err := h.Service.GetMyApotek(adminID)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	c.JSON(200, apotek)
}

// @Summary Mencari Apotek Terdekat
// @Description Cari apotek dalam radius tertentu menggunakan rumus Haversine.
// @Tags Apotek
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param radius query number false "Radius in KM (default 5)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} map[string]interface{} "data: []domain.Apotek, meta: object"
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
