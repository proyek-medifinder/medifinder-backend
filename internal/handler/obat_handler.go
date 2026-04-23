package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type ObatHandler struct {
	Service *service.ObatService
}

// CreateObat godoc
// @Summary Tambahkan obat baru
// @Description Admin apotek menambahkan obat ke apotek miliknya
// @Tags Admin Obat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ObatRequest true "Data obat"
// @Success 201 {object} dto.APIResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /admin/obat [post]
func (h *ObatHandler) Create(c *gin.Context) {
	adminID := c.GetString("user_id")
	fmt.Println("ADMIN ID FROM TOKEN:", adminID)

	var req struct {
		Nama  string  `json:"nama"`
		Stok  int     `json:"stok"`
		Harga float64 `json:"harga"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.Create(adminID, req.Nama, req.Stok, req.Harga)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "obat created"})
}

// Update godoc
// @Summary Update data obat
// @Description Update informasi nama, stok, atau harga obat oleh admin apotek
// @Tags Admin Obat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Obat ID"
// @Param request body dto.ObatRequest true "Data update obat"
// @Success 200 {object} map[string]interface{} "message: obat updated"
// @Failure 400 {object} map[string]interface{} "error: invalid input / error message"
// @Router /admin/obat/{id} [put]
func (h *ObatHandler) Update(c *gin.Context) {
	adminID := c.GetString("user_id")
	obatID := c.Param("id")

	var req struct {
		Nama  string  `json:"nama"`
		Stok  int     `json:"stok"`
		Harga float64 `json:"harga"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.Update(adminID, obatID, req.Nama, req.Stok, req.Harga)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "obat updated"})
}

// @Summary Hapus obat
// @Tags Admin Obat
// @Router /admin/obat/:id [delete]
func (h *ObatHandler) Delete(c *gin.Context) {
	adminID := c.GetString("user_id")
	obatID := c.Param("id")

	err := h.Service.Delete(adminID, obatID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "obat deleted"})
}

// GetByApotekPublic godoc
// @Summary Lihat daftar obat di apotek tertentu (Public)
// @Tags Obat
// @Param id path string true "Apotek ID"
// @Param name query string false "Search by Name"  <-- Tambahin baris dokumentasi ini buat Swagger
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} map[string]interface{}
// @Router /apotek/{id}/obat [get]
func (h *ObatHandler) GetByApotekPublic(c *gin.Context) {
	apotekID := c.Param("id") // Ambil :id dari URL

	// Ambil query params untuk pagination
	name := c.Query("name")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	// Konversi string ke int
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Panggil Service
	obat, total, err := h.Service.GetPublicByApotek(apotekID, name, limit, offset)
	if err != nil {
		// NAH DISINI: Kalau lu return 404 pas err != nil,
		// padahal ID-nya bener tapi datanya kosong, ini yang bikin bingung.
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Response Sukses
	c.JSON(http.StatusOK, gin.H{
		"data":  obat,
		"total": total,
	})
}

// GetMyObat godoc
// @Summary Lihat stok obat saya (Admin Only)
// @Tags Admin Obat
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /admin/obat [get]
func (h *ObatHandler) GetMyObat(c *gin.Context) {
	adminID := c.GetString("user_id")

	obat, err := h.Service.GetMyObat(adminID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": obat})
}
