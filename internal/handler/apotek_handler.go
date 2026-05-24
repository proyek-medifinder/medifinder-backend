package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/helper"
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

// Create godoc
// @Summary      Tambah Apotek Baru
// @Description   Membuat data apotek baru sekaligus upload foto profil
// @Tags         Apotek
// @Accept       multipart/form-data
// @Produce      json
// @Param        nama       formData  string  true  "Nama Apotek"
// @Param        alamat     formData  string  true  "Alamat Lengkap"
// @Param        latitude   formData  number  true  "Latitude"
// @Param        longitude  formData  number  true  "Longitude"
// @Param        jam_buka   formData  string  false "Format HH:mm"
// @Param        jam_tutup  formData  string  false "Format HH:mm"
// @Param        photo      formData  file    false "File Foto (.jpg, .png)"
// @Success      201  {object}  map[string]interface{}
// @Router       /admin/apotek [post]
// @Security     Bearer
func (h *ApotekHandler) Create(c *gin.Context) {

	nama := c.PostForm("nama")
	alamat := c.PostForm("alamat")
	lat, _ := strconv.ParseFloat(c.PostForm("latitude"), 64)
	lng, _ := strconv.ParseFloat(c.PostForm("longitude"), 64)
	jamBuka := c.PostForm("jam_buka")
	jamTutup := c.PostForm("jam_tutup")
	adminID := c.GetString("user_id")

	var photo_url *string
	file, err := c.FormFile("photo")

	if err == nil {
		src, err := file.Open()
		if err == nil {
			defer src.Close()
			url, err := helper.UploadToCloudinary(src) // Pakai helper yang udah kita bikin
			if err == nil {
				photo_url = &url
			} else {
				c.JSON(500, gin.H{"error": "Gagal upload ke Cloud: " + err.Error()})
				return
			}
		}
	}

	if photo_url != nil {
		fmt.Printf("DEBUG HANDLER: URL ada, isinya: %s\n", *photo_url)
	} else {
		fmt.Println("DEBUG HANDLER: photo_url MASIH NIL!")
	}

	err = h.Service.Create(adminID, nama, alamat, lat, lng, jamBuka, jamTutup, photo_url)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Apotek berhasil dibuat"})
}

// GetMyApotek godoc
// @Summary Lihat apotek milik admin
// @Description Mengambil data apotek yang dimiliki oleh admin yang sedang login
// @Tags Admin Apotek
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "data apotek"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Failure 404 {object} map[string]interface{} "apotek tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "internal error"
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

// UpdateMyApotek godoc
// @Summary Update Profil Apotek Sendiri
// @Description Update informasi profil apotek termasuk jam operasional dan kontak
// @Tags Apotek
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateApotekRequest true "Data update apotek"
// @Success 200 {object} map[string]interface{} "message: profil apotek berhasil diupdate"
// @Failure 400 {object} map[string]interface{} "error: format data tidak valid / pesan error lainnya"
// @Router /admin/apotek [put]
func (h *ApotekHandler) UpdateMyApotek(c *gin.Context) {
	adminID := c.GetString("user_id")

	var req dto.UpdateApotekRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "format data tidak valid"})
		return
	}

	// Panggil service dengan parameter lengkap
	err := h.Service.Update(adminID, req.Nama, req.Alamat, req.Latitude, req.Longitude, req.JamBuka, req.JamTutup, req.PhoneNumber, req.Deskripsi, req.PhotoURL)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "profil apotek berhasil diupdate"})
}

// SearchNearby godoc
// @Summary Cari apotek terdekat
// @Description Mengambil daftar apotek berdasarkan lokasi user (latitude & longitude)
// @Tags Apotek
// @Accept json
// @Produce json
// @Param lat query number true "Latitude user"
// @Param lng query number true "Longitude user"
// @Param radius query number false "Radius pencarian dalam km (default 5, max 50)"
// @Param page query int false "Page number"
// @Param limit query int false "Limit data"
// @Success 200 {object} map[string]interface{} "list apotek + pagination"
// @Failure 400 {object} map[string]interface{} "invalid parameter"
// @Failure 500 {object} map[string]interface{} "internal error"
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

func (h *ApotekHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	apotek, err := h.Service.GetByID(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Apotek kaga ada"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Mantap dapet",
		"data":    apotek,
	})
}

// UpdatePhoto godoc
// @Summary      Update Foto Apotek
// @Description   Upload foto profil untuk apotek milik admin yang sedang login
// @Tags         Apotek
// @Accept       multipart/form-data
// @Produce      json
// @Param        photo  formData  file  true  "File Foto Apotek (.jpg, .png)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Security     Bearer
// @Router       /admin/foto [put]
func (h *ApotekHandler) UpdatePhoto(c *gin.Context) {
	adminID := c.GetString("user_id")

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(400, gin.H{"error": "Foto tidak ditemukan di request"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal membuka file"})
		return
	}
	defer src.Close()

	// Upload ke Cloudinary
	url, err := helper.UploadToCloudinary(src)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal upload ke Cloud: " + err.Error()})
		return
	}

	// Update ke Database (Simpan URL-nya saja)
	err = h.Service.UpdateImage(adminID, url)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "Foto apotek berhasil diupdate",
		"url":     url,
	})
}

// SuperAdminIndex godoc
// @Summary      Daftar Semua Apotek (Super Admin)
// @Description  Mengambil semua daftar apotek yang terdaftar di sistem untuk kebutuhan monitoring
// @Tags         SuperAdmin Apotek
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Halaman"
// @Param        limit query int false "Limit data"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /superadmin/apotek [get]
func (h *ApotekHandler) SuperAdminIndex(c *gin.Context) {
	// Gunakan helper pagination yang sudah kamu buat sebelumnya
	page, limit, _ := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetAllForSuperAdmin(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar apotek"})
		return
	}

	totalPage := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": totalPage,
		},
	})
}

// MobileIndex godoc
// @Summary      Daftar Semua Apotek (Mobile)
// @Description  Mengambil semua daftar apotek untuk kebutuhan aplikasi mobile tanpa login
// @Tags         Mobile Apotek
// @Produce      json
// @Param        page query int false "Halaman"
// @Param        limit query int false "Limit data"
// @Success      200 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{} "Internal server error"
// @Router       /mobile/apotek [get]
func (h *ApotekHandler) MobileIndex(c *gin.Context) {
	page, limit, _ := utils.GetPaginationAdvanced(c)

	data, total, err := h.Service.GetAllForSuperAdmin(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar apotek"})
		return
	}

	totalPage := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data": data,
		"meta": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_page": totalPage,
		},
	})
}
