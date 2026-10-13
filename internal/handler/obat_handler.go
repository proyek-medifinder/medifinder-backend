package handler

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/helper"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type ObatHandler struct {
	Service *service.ObatService
}

// helper untuk mengambil file upload dari berbagai kemungkinan field name
func getUploadedImage(c *gin.Context) (*multipart.FileHeader, error) {
	fields := []string{"image", "foto", "photo", "file", "image_url"}
	for _, field := range fields {
		file, err := c.FormFile(field)
		if err == nil {
			return file, nil
		}
	}
	return nil, http.ErrMissingFile
}

// Create godoc
// @Summary Tambahkan obat baru (dengan opsi upload foto ke Cloudinary)
// @Description Admin apotek menambahkan obat ke apotek miliknya (Mendukung multipart/form-data dan JSON)
// @Tags Admin Obat
// @Security BearerAuth
// @Accept multipart/form-data,json
// @Produce json
// @Param nama formData string true "Nama obat"
// @Param stok formData int true "Stok obat"
// @Param harga formData number true "Harga obat"
// @Param deskripsi formData string false "Deskripsi obat"
// @Param image formData file false "Foto obat"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /admin/obat [post]
func (h *ObatHandler) Create(c *gin.Context) {
	adminID := c.GetString("user_id")

	var (
		nama      string
		stok      int
		harga     float64
		deskripsi *string
		imageURL  *string
	)

	contentType := c.ContentType()

	if strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "application/x-www-form-urlencoded") {
		nama = strings.TrimSpace(c.PostForm("nama"))
		stok, _ = strconv.Atoi(c.PostForm("stok"))
		harga, _ = strconv.ParseFloat(c.PostForm("harga"), 64)

		if desk := strings.TrimSpace(c.PostForm("deskripsi")); desk != "" {
			deskripsi = &desk
		}
		if imgStr := strings.TrimSpace(c.PostForm("image_url")); imgStr != "" {
			imageURL = &imgStr
		}

		// Cek apakah ada file foto yang di-upload
		if file, err := getUploadedImage(c); err == nil {
			src, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file: " + err.Error()})
				return
			}
			defer src.Close()

			url, err := helper.UploadToCloudinary(src)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke Cloudinary: " + err.Error()})
				return
			}
			imageURL = &url
		}
	} else {
		// JSON payload
		var req dto.ObatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format input JSON tidak valid: " + err.Error()})
			return
		}
		nama = strings.TrimSpace(req.Nama)
		stok = req.Stok
		harga = req.Harga
		deskripsi = req.Deskripsi
		imageURL = req.ImageURL
	}

	if nama == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama obat wajib diisi"})
		return
	}

	obat, err := h.Service.Create(adminID, nama, deskripsi, stok, harga, imageURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "obat created",
		"data":    obat,
	})
}

// Update godoc
// @Summary Update data obat (dengan opsi upload foto baru ke Cloudinary)
// @Description Update informasi nama, stok, harga, deskripsi, atau foto obat oleh admin apotek
// @Tags Admin Obat
// @Security BearerAuth
// @Accept multipart/form-data,json
// @Produce json
// @Param id path string true "Obat ID"
// @Param nama formData string false "Nama obat"
// @Param stok formData int false "Stok obat"
// @Param harga formData number false "Harga obat"
// @Param deskripsi formData string false "Deskripsi obat"
// @Param image formData file false "Foto baru obat"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /admin/obat/{id} [put]
func (h *ObatHandler) Update(c *gin.Context) {
	adminID := c.GetString("user_id")
	obatID := c.Param("id")

	var (
		nama      string
		stok      int
		harga     float64
		deskripsi *string
		imageURL  *string
	)

	contentType := c.ContentType()

	if strings.Contains(contentType, "multipart/form-data") || strings.Contains(contentType, "application/x-www-form-urlencoded") {
		nama = strings.TrimSpace(c.PostForm("nama"))
		stok, _ = strconv.Atoi(c.PostForm("stok"))
		harga, _ = strconv.ParseFloat(c.PostForm("harga"), 64)

		if desk := strings.TrimSpace(c.PostForm("deskripsi")); desk != "" {
			deskripsi = &desk
		}
		if imgStr := strings.TrimSpace(c.PostForm("image_url")); imgStr != "" {
			imageURL = &imgStr
		}

		// Cek upload foto baru
		if file, err := getUploadedImage(c); err == nil {
			src, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file: " + err.Error()})
				return
			}
			defer src.Close()

			url, err := helper.UploadToCloudinary(src)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke Cloudinary: " + err.Error()})
				return
			}
			imageURL = &url
		}
	} else {
		// JSON payload
		var req dto.ObatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format input JSON tidak valid: " + err.Error()})
			return
		}
		nama = strings.TrimSpace(req.Nama)
		stok = req.Stok
		harga = req.Harga
		deskripsi = req.Deskripsi
		imageURL = req.ImageURL
	}

	if nama == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama obat wajib diisi"})
		return
	}

	obat, err := h.Service.Update(adminID, obatID, nama, deskripsi, stok, harga, imageURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "obat updated",
		"data":    obat,
	})
}

// UpdateImage godoc
// @Summary Upload atau ubah foto obat ke Cloudinary (Admin Only)
// @Description Mengunggah gambar baru untuk obat tertentu ke Cloudinary dan menyimpan image_url ke database
// @Tags Admin Obat
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Obat ID"
// @Param image formData file true "File Foto Obat"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /admin/obat/{id}/foto [put]
func (h *ObatHandler) UpdateImage(c *gin.Context) {
	adminID := c.GetString("user_id")
	obatID := c.Param("id")

	file, err := getUploadedImage(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File foto obat tidak ditemukan di request"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuka file: " + err.Error()})
		return
	}
	defer src.Close()

	// Upload ke Cloudinary
	url, err := helper.UploadToCloudinary(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal upload ke Cloudinary: " + err.Error()})
		return
	}

	// Simpan URL ke Database
	err = h.Service.UpdateImage(adminID, obatID, url)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Foto obat berhasil diupdate",
		"image_url": url,
		"data": gin.H{
			"image_url": url,
		},
	})
}

// GetMyObat godoc
// @Summary Tampilkan semua obat apotek saya (Admin Only)
// @Tags Admin Obat
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /admin/obat [get]
func (h *ObatHandler) GetMyObat(c *gin.Context) {
	adminID := c.GetString("user_id")

	obat, err := h.Service.GetMyObat(adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": obat})
}

// GetMyObatByID godoc
// @Summary Tampilkan detail satu obat milik apotek saya (Admin Only)
// @Tags Admin Obat
// @Security BearerAuth
// @Param id path string true "Obat ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /admin/obat/{id} [get]
func (h *ObatHandler) GetMyObatByID(c *gin.Context) {
	adminID := c.GetString("user_id")
	obatID := c.Param("id")

	obat, err := h.Service.GetMyObatByID(adminID, obatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": obat})
}

// GetByIDPublic godoc
// @Summary Tampilkan detail satu obat (Public)
// @Tags Obat
// @Param id path string true "Obat ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /obat/{id} [get]
func (h *ObatHandler) GetByIDPublic(c *gin.Context) {
	obatID := c.Param("id")

	obat, err := h.Service.GetByID(obatID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Obat tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": obat})
}

// GetByApotekPublic godoc
// @Summary Lihat daftar obat di apotek tertentu (Public)
// @Tags Obat
// @Param id path string true "Apotek ID"
// @Param name query string false "Search by Name"
// @Param offset query int false "Offset"
// @Param limit query int false "Limit"
// @Success 200 {object} map[string]interface{}
// @Router /apotek/{id}/obat [get]
func (h *ObatHandler) GetByApotekPublic(c *gin.Context) {
	apotekID := c.Param("id")

	name := c.Query("name")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	obat, total, err := h.Service.GetPublicByApotek(apotekID, name, limit, offset)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  obat,
		"total": total,
	})
}

// Delete godoc
// @Summary Hapus obat
// @Description Menghapus data obat milik admin apotek
// @Tags Admin Obat
// @Security BearerAuth
// @Param id path string true "Obat ID"
// @Success 200 {object} map[string]interface{} "obat deleted"
// @Failure 400 {object} map[string]interface{} "error message"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Router /admin/obat/{id} [delete]
func (h *ObatHandler) Delete(c *gin.Context) {
	adminID := c.GetString("user_id")
	obatID := c.Param("id")

	err := h.Service.Delete(adminID, obatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "obat deleted"})
}

