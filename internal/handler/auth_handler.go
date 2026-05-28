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
// @Description Autentikasi user dan mengembalikan JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Email & Password"
// @Success 200 {object} dto.AuthResponse "token, user info"
// @Failure 400 {object} dto.ErrorResponse "invalid input"
// @Failure 401 {object} dto.ErrorResponse "invalid credentials"
// @Failure 500 {object} dto.ErrorResponse "internal error"
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

// ForgotPassword godoc
// @Summary      Request Reset Password
// @Description  Mengirimkan link/token ke email admin untuk melakukan reset password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ForgotPasswordRequest true "Email Admin"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format email tidak valid"})
		return
	}

	h.Service.ForgotPassword(req.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Jika email terdaftar, link reset telah dikirim",
	})
}

// ResetPassword godoc
// @Summary      Reset Password via Token
// @Description  Membuat password baru menggunakan token yang dikirimkan via email
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.ResetPasswordRequest true "Token dari Email & Password Baru"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Router       /reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token dan password baru (minimal 6 karakter) wajib diisi"})
		return
	}

	err := h.Service.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil direset"})
}

// ChangePassword godoc
// @Summary      Ganti Password (Login Required)
// @Description  Mengganti password admin yang sedang aktif (harus tahu password lama)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body dto.ChangePasswordRequest true "Password Lama dan Password Baru"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Router       /change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {

	var req dto.ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password lama dan baru wajib diisi"})
		return
	}

	// AMAN: Ambil ID dari token JWT yang lagi login, bukan dari body request
	userIDClaim, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := uuid.Parse(userIDClaim.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID format"})
		return
	}

	err = h.Service.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diganti"})
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

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Fungsi ini biasanya buat nangkep token dari query string
	// atau nanganin redirect setelah Google selesai autentikasi
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code tidak ditemukan"})
		return
	}

	// Panggil service untuk menukar code dengan profile Google
	// Tergantung implementasi lu, bisa redirect ke frontend bawa token
	c.JSON(http.StatusOK, gin.H{"message": "Callback diterima", "code": code})
}
