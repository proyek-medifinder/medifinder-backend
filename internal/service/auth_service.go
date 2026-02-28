package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"

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
	if user.Status == "suspended" {
		return nil, errors.New("akun anda telah dinonaktifkan oleh Super Admin")
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

func (s *AuthService) sendEmailGomail(toEmail string, token string) {
	m := gomail.NewMessage()
	m.SetHeader("From", "cs.medifinder@gmail.com")
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Reset Password Akun Medifinder")

	// Perbaikan os.Getenv
	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	htmlBody := fmt.Sprintf(`
		<h3>Halo,</h3>
		<p>Kami menerima permintaan untuk mereset password akun Medifinder Anda.</p>
		<p>Silakan klik link di bawah ini untuk mereset password Anda:</p>
		<a href="%s" style="background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Reset Password</a>
		<br><br>
		<p>Jika Anda tidak pernah meminta reset password, abaikan email ini.</p>
	`, resetLink)

	m.SetBody("text/html", htmlBody)

	// Ambil port dari .env, default ke 465 jika tidak diset
	portStr := os.Getenv("SMTP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port == 0 {
		port = 465
	}

	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_EMAIL"),
		os.Getenv("SMTP_PASS"),
	)

	if err := d.DialAndSend(m); err != nil {
		log.Println("❌ GAGAL kirim email ke", toEmail, "Error:", err)
	} else {
		log.Println("✅ SUKSES kirim email reset password ke:", toEmail)
	}
}

func (s *AuthService) ForgotPassword(email string) {
	token, _ := utils.GenerateResetToken(email)

	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	resetLink := fmt.Sprintf(
		"%s/reset-password?token=%s",
		baseURL,
		token,
	)

	body := fmt.Sprintf(`
	<h2>Reset Password</h2>
	<p>Klik link di bawah:</p>
	<a href="%s">Reset Password</a>
	<p>Link berlaku 15 menit.</p>
	`, resetLink)

	utils.SendEmail(email, "Reset Password", body)
}

func (s *AuthService) ResetPassword(token, newPassword string) error {
	email, err := utils.VerifyResetToken(token)
	if err != nil {
		return fmt.Errorf("token tidak valid atau sudah kadaluarsa")
	}

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gagal memproses password baru")
	}

	err = s.UserRepo.UpdatePassword(user.ID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("gagal menyimpan password ke database")
	}

	log.Println("✅ Password berhasil direset untuk email:", email)

	s.sendPasswordChangedEmail(email)

	return nil
}

func (s *AuthService) ChangePassword(email, newPassword string) error {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gagal memproses password baru")
	}

	err = s.UserRepo.UpdatePassword(user.ID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("gagal menyimpan password ke database")
	}

	log.Println("✅ Password berhasil diganti untuk email:", email)

	s.sendPasswordChangedEmail(email)

	return nil
}

func (s *AuthService) sendPasswordChangedEmail(email string) {
	body := `
	<h3>Password berhasil diubah</h3>
	<p>Jika ini bukan Anda, segera hubungi admin atau ubah password Anda kembali.</p>
	`
	utils.SendEmail(email, "Password Berhasil Diubah", body)
}

func (s *AuthService) RegisterAdmin(name, email, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: string(hashed),
		RoleID:   2,
		Status:   "pending",
	}

	return s.UserRepo.Create(user)
}
