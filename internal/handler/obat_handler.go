package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"github.com/sasaefulanwar/medifinder/internal/utils"
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
		Nama  string `json:"nama"`
		Stok  int    `json:"stok"`
		Harga int64  `json:"harga"`
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

// @Summary Update data obat
// @Tags Admin Obat
// @Router /admin/obat/:id [put]
func (h *ObatHandler) Update(c *gin.Context) {
	adminID := c.GetString("user_id")
	obatID := c.Param("id")

	var req struct {
		Nama  string `json:"nama"`
		Stok  int    `json:"stok"`
		Harga int64  `json:"harga"`
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

func (h *ObatHandler) GetByApotekPublic(c *gin.Context) {

	apotekID := c.Param("id")

	limit, offset := utils.GetPagination(c)

	obat, total, err := h.Service.GetPublicByApotek(apotekID, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch data"})
		return
	}

	c.JSON(200, gin.H{
		"data":  obat,
		"total": total,
		"limit": limit,
		"page":  (offset / limit) + 1,
	})
}

func (h *ObatHandler) GetMyObat(c *gin.Context) {
	adminID := c.GetString("user_id")

	obat, err := h.Service.GetMyObat(adminID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"data": obat})
}
