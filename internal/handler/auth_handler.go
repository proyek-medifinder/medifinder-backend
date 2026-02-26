package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/service"
)

type AuthHandler struct {
	Service *service.AuthService
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

// ForgotPassword godoc
// @Summary Request reset password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Email user"
// @Success 200 {object} map[string]interface{}
// @Router /forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.ForgotPassword(req); err != nil { // Pakai h.Service
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instruksi reset password telah dikirim ke email"})
}

// ResetPassword godoc
// @Summary Reset password dengan token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Token & password baru"
// @Success 200 {object} map[string]interface{}
// @Router /reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Service.ResetPassword(req); err != nil { // Pakai h.Service
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah, silakan login kembali"})
}

// ChangePassword godoc
// @Summary Ubah password user yang sedang login
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Password lama dan baru (old_password, new_password)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Ambil ID dari token JWT (Middleware lu udah otomatis nyimpen ini di context)
	userID := c.GetString("user_id")

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Format input tidak valid"})
		return
	}

	if err := h.Service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Password berhasil diubah"})
}

// @Summary Registrasi Mandiri Admin Apotek
// @Tags Auth
// @Produce json
// @Router /register-admin [post]
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Format data tidak valid"})
		return
	}

	err := h.Service.RegisterAdmin(req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal mendaftarkan admin apotek: " + err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "Registrasi Admin Apotek berhasil. Silakan tunggu verifikasi dari Super Admin sebelum dapat login.",
	})
}
