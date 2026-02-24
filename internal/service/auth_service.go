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
		Status:   "approved",
	}

	return s.UserRepo.Create(user)
}

func (s *AuthService) Login(email, password string) (*dto.AuthResponse, error) {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email/password")
	}

	if user.Status == "pending" {
		return nil, errors.New("akun anda sedang dalam tahap verifikasi oleh Super Admin")
	}
	if user.Status == "rejected" {
		return nil, errors.New("pendaftaran akun anda ditolak")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid email/password")
	}

	_ = s.UserRepo.UpdateLastLogin(user.ID)

	role := "user"
	if user.RoleID == 2 {
		role = "admin_apotek"
	}
	if user.RoleID == 3 {
		role = "super_admin"
	}

	token, err := utils.GenerateJWT(user.ID, role)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		Name:  user.Name,
		Email: user.Email,
		Role:  role,
	}, nil
}

func (s *AuthService) GoogleLogin(googleToken string) (*dto.AuthResponse, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		clientID = "539021546127-kj6icorqjdrouo9n31tla5r2tcl90e7r.apps.googleusercontent.com"
	}

	payload, err := idtoken.Validate(context.Background(), googleToken, clientID)
	if err != nil {
		return nil, errors.New("token google tidak valid atau expired")
	}

	email := payload.Claims["email"].(string)
	name := payload.Claims["name"].(string)
	googleID := payload.Subject

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		randomPassword := generateRandomToken()[:10]
		hashed, _ := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)

		newUser := &domain.User{
			ID:       uuid.New(),
			Name:     name,
			Email:    email,
			Password: string(hashed),
			RoleID:   1,
			GoogleID: &googleID,
			Status:   "approved",
		}

		err = s.UserRepo.Create(newUser)
		if err != nil {
			return nil, err
		}
		user = newUser
	}

	if user.Status == "pending" {
		return nil, errors.New("akun anda sedang dalam tahap verifikasi oleh Super Admin")
	}
	if user.Status == "rejected" {
		return nil, errors.New("pendaftaran akun anda ditolak")
	}

	_ = s.UserRepo.UpdateLastLogin(user.ID)

	role := "user"
	if user.RoleID == 2 {
		role = "admin_apotek"
	} else if user.RoleID == 3 {
		role = "super_admin"
	}

	token, err := utils.GenerateJWT(user.ID, role)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		Name:  user.Name,
		Email: user.Email,
		Role:  role,
	}, nil
}

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

// ================ FITUR REGISTER ADMIN APOTEK (SUPERADMIN) =================
func (s *AuthService) RegisterAdmin(name, email, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: string(hashed),
		RoleID:   2,         // Role 2 = Admin Apotek
		Status:   "pending", // Default status: Belum Terverifikasi
	}

	return s.UserRepo.Create(user)
}
