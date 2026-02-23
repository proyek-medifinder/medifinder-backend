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

	token, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"token": token})
}

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
