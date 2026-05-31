package controllers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type TokenRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

func GoogleLoginAndProfile(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token wajib dikirim"})
		return
	}

	ctx := context.Background()

	// 1. Inisialisasi Google OAuth2 Service
	oauth2Service, err := oauth2.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal inisialisasi layanan Google"})
		return
	}

	// 2. Validasi Token ke Google (Memastikan token asli & belum expired)
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(req.IDToken)

	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token Google tidak valid atau sudah kedaluwarsa"})
		return
	}

	// 3. TRIK JINJA YANG BENER: Decode Payload JWT secara lokal buat ambil Name & Picture
	var name, picture string
	parts := strings.Split(req.IDToken, ".")
	if len(parts) == 3 {
		// Payload JWT ada di bagian tengah (index 1)
		payloadSegment := parts[1]
		
		// Handle base64 padding jika kurang
		if rem := len(payloadSegment) % 4; rem != 0 {
			payloadSegment += strings.Repeat("=", 4-rem)
		}
		
		payloadBytes, err := base64.URLEncoding.DecodeString(payloadSegment)
		if err == nil {
			var claims map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &claims); err == nil {
				name, _ = claims["name"].(string)
				picture, _ = claims["picture"].(string)
			}
		}
	}

	// 4. Susun data profil user
	userProfile := gin.H{
		"google_id": tokenInfo.UserId,
		"email":     tokenInfo.Email,
		"name":      name,    // Sekarang dapet tanpa error compiler!
		"picture":   picture, // Foto profil juga dapet!
	}

	// 5. [LOGIC database medifinder lo di sini]

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Autentikasi Google berhasil",
		"data":    userProfile,
	})
}