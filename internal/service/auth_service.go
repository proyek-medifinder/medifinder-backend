package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/repository"
	"github.com/sasaefulanwar/medifinder/internal/utils"
	"google.golang.org/api/idtoken"
)

type AuthService struct {
	UserRepo *repository.UserRepository
}

func (s *AuthService) Register(name, email, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: string(hashed),
		RoleID:   1,
	}

	return s.UserRepo.Create(user)
}

func (s *AuthService) Login(email, password string) (string, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid email/password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email/password")
	}

	role := "user"
	if user.RoleID == 2 {
		role = "admin_apotek"
	}
	if user.RoleID == 3 {
		role = "super_admin"
	}

	return utils.GenerateJWT(user.ID, role)
}

// ================= FITUR GOOGLE OAUTH2 =================

func (s *AuthService) GoogleLogin(googleToken string) (string, error) {
	// 1. Ambil Client ID dari environment (.env) sesuai SRS NFR-16
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		clientID = "407408718192.apps.googleusercontent.com"
	}

	// 2. Verifikasi token ke server Google
	payload, err := idtoken.Validate(context.Background(), googleToken, clientID)
	if err != nil {
		return "", errors.New("token google tidak valid atau expired")
	}

	// 3. Ambil data dari payload Google
	email := payload.Claims["email"].(string)
	name := payload.Claims["name"].(string)
	googleID := payload.Subject

	// 4. Cek apakah email udah ada di database
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		// USER BELUM ADA -> AUTO REGISTER
		// Karena kolom password di DB lo NOT NULL, kita buatin password acak
		randomPassword := generateRandomToken()[:10]
		hashed, _ := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)

		newUser := &domain.User{
			ID:       uuid.New(),
			Name:     name,
			Email:    email,
			Password: string(hashed),
			RoleID:   1, // Role default: User
			GoogleID: &googleID,
		}

		err = s.UserRepo.Create(newUser)
		if err != nil {
			return "", err
		}
		user = newUser // Pake data user baru buat generate JWT
	}

	// 5. Generate JWT Medifinder buat login
	role := "user"
	if user.RoleID == 2 {
		role = "admin_apotek"
	} else if user.RoleID == 3 {
		role = "super_admin"
	}

	return utils.GenerateJWT(user.ID, role)
}

// ================= FITUR BARU RESET PASSWORD =================

func generateRandomToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *AuthService) ForgotPassword(req dto.ForgotPasswordRequest) error {
	user, err := s.UserRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		return errors.New("email tidak terdaftar")
	}

	token := generateRandomToken()
	expiry := time.Now().Add(1 * time.Hour)

	err = s.UserRepo.UpdateResetToken(req.Email, token, expiry)
	if err != nil {
		return err
	}

	log.Printf("BOHONG-BOHONGAN KIRIM EMAIL: Klik link ini untuk reset password -> http://localhost:8080/reset-password?token=%s\n", token)
	return nil
}

func (s *AuthService) ResetPassword(req dto.ResetPasswordRequest) error {
	user, err := s.UserRepo.FindByResetToken(req.Token)
	if err != nil || user == nil {
		return errors.New("token tidak valid atau tidak ditemukan")
	}

	if time.Now().After(*user.ResetTokenExpiry) {
		return errors.New("token sudah kedaluwarsa")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.UserRepo.ClearResetToken(user.ID, string(hashedPassword))
}
