package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"github.com/sasaefulanwar/medifinder/internal/utils"
)

type SuperAdminHandler struct {
	Service *service.SuperAdminService
}

// ListAdmin godoc
// @Summary Daftar semua admin apotek
// @Tags SuperAdmin
// @Security BearerAuth
// @Router /superadmin/admin [get]
func (h *SuperAdminHandler) ListAdmin(c *gin.Context) {

	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.ListAdmin(limit, offset)
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

// CreateAdmin godoc
// @Summary Buat akun admin apotek baru
// @Tags SuperAdmin
// @Security BearerAuth
// @Param request body dto.AdminRequest true "Data admin"
// @Router /superadmin/admin [post]
func (h *SuperAdminHandler) CreateAdmin(c *gin.Context) {

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.CreateAdmin(req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "admin created"})
}

func (h *SuperAdminHandler) UpdateAdmin(c *gin.Context) {

	id := c.Param("id")

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.UpdateAdmin(id, req.Name, req.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "admin updated"})
}

func (h *SuperAdminHandler) DeleteAdmin(c *gin.Context) {

	id := c.Param("id")

	err := h.Service.DeleteAdmin(id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "admin deleted"})
}

func (h *SuperAdminHandler) GetPendingAdmins(c *gin.Context) {
	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetPendingAdmins(limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"data": data,
		"meta": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func (h *SuperAdminHandler) VerifyAdmin(c *gin.Context) {
	// Ambil ID Super Admin yang lagi login dari token JWT
	superAdminID := c.GetString("user_id")

	var req struct {
		AdminID string `json:"admin_id" binding:"required"`
		Action  string `json:"action" binding:"required"` // 'approved' atau 'rejected'
		Notes   string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "format data tidak valid"})
		return
	}

	err := h.Service.VerifyAdmin(req.AdminID, superAdminID, req.Action, req.Notes)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Berhasil memverifikasi pengajuan admin apotek"})
}
