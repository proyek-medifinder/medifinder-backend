package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/service"
	"github.com/sasaefulanwar/medifinder/internal/utils"
)

type TransaksiHandler struct {
	Service *service.TransaksiService
}

func (h *TransaksiHandler) UserHistory(c *gin.Context) {

	userID := c.GetString("user_id")
	status := c.Query("status")

	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetUserHistory(userID, status, limit, offset)
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

func (h *TransaksiHandler) Detail(c *gin.Context) {

	id := c.Param("id")

	data, err := h.Service.GetDetail(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *TransaksiHandler) AdminHistory(c *gin.Context) {

	adminID := c.GetString("user_id")
	status := c.Query("status")

	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetAdminHistory(adminID, status, limit, offset)
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

func (h *TransaksiHandler) SuperAdminHistory(c *gin.Context) {

	status := c.Query("status")
	page, limit, offset := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetAllHistory(status, limit, offset)
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
