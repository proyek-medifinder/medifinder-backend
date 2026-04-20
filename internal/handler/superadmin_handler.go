package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/dto"
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

// GetPendingAdmins godoc
// @Summary Ambil daftar admin apotek yang masih pending (butuh verifikasi)
// @Tags SuperAdmin
// @Security BearerAuth
// @Produce json
// @Param page query int false "Halaman (default 1)" default(1)
// @Param limit query int false "Jumlah data per halaman (default 10)" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /superadmin/pengajuan [get]
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

// VerifyAdmin godoc
// @Summary Verifikasi pendaftaran admin apotek (Approve / Reject)
// @Tags SuperAdmin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.VerifyAdminRequest true "Data Verifikasi (Action: approved/rejected)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /superadmin/verifikasi [post]
func (h *SuperAdminHandler) VerifyAdmin(c *gin.Context) {
	superAdminID := c.GetString("user_id")

	// GANTI INLINE STRUCT JADI DTO
	var req dto.VerifyAdminRequest

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

// ChangeAdminStatus godoc
// @Summary Ubah status admin apotek (Suspend / Activate)
// @Tags SuperAdmin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "ID Admin Apotek"
// @Param request body dto.ChangeAdminStatusRequest true "Status baru: 'approved' atau 'suspended'"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /superadmin/admin/{id}/status [patch]
func (h *SuperAdminHandler) ChangeAdminStatus(c *gin.Context) {
	adminID := c.Param("id")

	// GANTI INLINE STRUCT JADI DTO
	var req dto.ChangeAdminStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Format input tidak valid"})
		return
	}

	if err := h.Service.ChangeAdminStatus(adminID, req.Status); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Status admin berhasil diubah menjadi " + req.Status,
	})
}
