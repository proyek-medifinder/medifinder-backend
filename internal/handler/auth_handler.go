package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type AuthHandler struct {
	Service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{Service: s}
}

// Register godoc
// @Summary Registrasi user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register data"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	err := h.Service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "registered successfully"})
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login data"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	// Tangkap response lengkap dari Service
	res, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	// Kirim res yang berisi Token, Name, dan Role
	c.JSON(200, res)
}

// GoogleLogin godoc
// @Summary Login via Google
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.GoogleLoginRequest true "Google token"
// @Success 200 {object} map[string]interface{}
// @Router /google-login [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dto.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format token tidak valid"})
		return
	}

	// Tangkap response lengkap dari Service
	res, err := h.Service.GoogleLogin(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}

	c.ShouldBindJSON(&req)

	h.Service.ForgotPassword(req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Jika email terdaftar, link reset telah dikirim",
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	c.ShouldBindJSON(&req)

	err := h.Service.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		c.JSON(400, gin.H{"error": "token tidak valid"})
		return
	}

	c.JSON(200, gin.H{"message": "password berhasil direset"})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}

	c.ShouldBindJSON(&req)

	h.Service.ChangePassword(req.Email, req.NewPassword)

	c.JSON(200, gin.H{"message": "password berhasil diganti"})
}

// RegisterAdmin godoc
// @Summary Registrasi Mandiri Admin Apotek
// @Description Pendaftaran mandiri untuk Admin Apotek beserta data apoteknya
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterAdminRequest true "Data Registrasi Admin Apotek"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /register-admin [post]
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	// Pake DTO yang baru kita bikin
	var req dto.RegisterAdminRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak lengkap atau tidak valid: " + err.Error()})
		return
	}

	err := h.Service.RegisterAdmin(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendaftar: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pendaftaran Admin Apotek berhasil diajukan, silakan tunggu verifikasi."})
}

// GetMe godoc
// @Summary Ambil profil user yang sedang login
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Ambil user_id hasil ekstraksi JWT dari middleware
	userIDClaim, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse claim token ke string lalu ke UUID
	userIDStr, ok := userIDClaim.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "format token user_id tidak valid"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "format UUID tidak valid"})
		return
	}

	// Panggil service
	userData, err := h.Service.GetMe(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    userData,
	})
}
