package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/repository"
	"github.com/sasaefulanwar/medifinder/internal/utils"
	"google.golang.org/api/idtoken"
)

var (
	RoleUserUUID       = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	RoleAdminUUID      = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	RoleSuperAdminUUID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
)

type AuthService struct {
	UserRepo *repository.UserRepository
}

// +++++++++++++++++ AUTH SERVICES BELOW +++++++++++++++++

func (s *AuthService) Register(name, email, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: string(hashed),
		RoleID:   RoleUserUUID,
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
	if user.RoleID == RoleAdminUUID {
		role = "admin_apotek"
	} else if user.RoleID == RoleSuperAdminUUID {
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

	// BONUS SOLUSI: Ambil link foto profil dari claims Google
	var picture string
	if p, ok := payload.Claims["picture"].(string); ok {
		picture = p
	}

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		// ---- KONDISI 1: USER BARU ----
		randomPassword := generateRandomToken()[:10]
		hashed, _ := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)

		newUser := &domain.User{
			ID:             uuid.New(),
			Name:           name,
			Email:          email,
			Password:       string(hashed),
			RoleID:         RoleUserUUID,
			GoogleID:       &googleID,
			ProfilePicture: &picture, // SEKARANG FOTO PROFIL KE-RECORD SAAAT DAFTAR
			Status:         "approved",
		}

		err = s.UserRepo.Create(newUser)
		if err != nil {
			return nil, err
		}
		user = newUser

	} else {
		// ---- KONDISI 2: USER LAMA LINKING KE GOOGLE ----
		var butuhFetchUlang bool

		// 1. Cek & Update Google ID jika kosong
		if user.GoogleID == nil || *user.GoogleID == "" {
			errUpdate := s.UserRepo.UpdateGoogleID(user.ID, googleID)
			if errUpdate != nil {
				log.Println("Gagal update Google ID:", errUpdate)
			} else {
				butuhFetchUlang = true
			}
		}

		// 2. Ambil foto profil dari Google Claims
		var picture string
		if p, ok := payload.Claims["picture"].(string); ok {
			picture = p
		}

		// Cek & Update Profile Picture jika di DB kosong
		if picture != "" && (user.ProfilePicture == nil || *user.ProfilePicture == "") {
			errUpdatePic := s.UserRepo.UpdateProfilePicture(user.ID, picture)
			if errUpdatePic != nil {
				log.Println("Gagal update Profile Picture:", errUpdatePic)
			} else {
				user.ProfilePicture = &picture
				butuhFetchUlang = true
			}
		}

		if butuhFetchUlang {
			freshUser, errFetch := s.UserRepo.FindByEmail(user.Email)
			if errFetch == nil {
				user = freshUser
			}
		}
	}

	if user.Status == "pending" {
		return nil, errors.New("akun anda sedang dalam tahap verifikasi oleh Super Admin")
	}
	if user.Status == "rejected" {
		return nil, errors.New("pendaftaran akun anda ditolak")
	}

	_ = s.UserRepo.UpdateLastLogin(user.ID)

	role := "user"
	if user.RoleID == RoleAdminUUID {
		role = "admin_apotek"
	} else if user.RoleID == RoleSuperAdminUUID {
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

func (s *AuthService) GetMe(userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.UserRepo.GetUserProfile(userID)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	role := "user"
	if user.RoleID == RoleAdminUUID {
		role = "admin_apotek"
	} else if user.RoleID == RoleSuperAdminUUID {
		role = "super_admin"
	}

	return map[string]interface{}{
		"id":              user.ID,
		"name":            user.Name,
		"email":           user.Email,
		"role":            role,
		"status":          user.Status,
		"profile_picture": user.ProfilePicture,
		"google_id":       user.GoogleID,
	}, nil
}

// +++++++++++++++++ PASSWORD RELATED SERVICES BELOW +++++++++++++++++

func (s *AuthService) ForgotPassword(email string) {
	user, _ := s.UserRepo.FindByEmail(email)
	name := "User"
	if user != nil {
		name = user.Name
	}

	token, _ := utils.GenerateResetToken(email)

	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	data := struct {
		Name      string
		ResetLink string
	}{
		Name:      name,
		ResetLink: resetLink,
	}

	body, err := utils.ParseTemplate("templates/emails/reset_password.html", data)
	if err != nil {
		log.Println("❌ Gagal render template email:", err)
		body = "Klik link ini untuk reset password: " + resetLink
	}

	// 4. Kirim!
	utils.SendEmail(email, "Reset Password Akun Medifinder", body)
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

func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {

	email, currentHash, err := s.UserRepo.GetAuthDataByID(userID)
	if err != nil {
		return fmt.Errorf("user tidak ditemukan")
	}

	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword))
	if err != nil {
		return fmt.Errorf("password lama salah")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gagal memproses password baru")
	}

	err = s.UserRepo.UpdatePassword(userID, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("gagal menyimpan password ke database")
	}

	log.Println("✅ Password berhasil diganti untuk user ID:", userID)

	// 5. Kirim notifikasi pake email yang ditarik tadi
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

// +++++++++++++++++ ADMIN REGISTRATION SERVICE BELOW +++++++++++++++++

func (s *AuthService) RegisterAdmin(req dto.RegisterAdminRequest) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	userID := uuid.New()

	user := &domain.User{
		ID:       userID,
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
		RoleID:   RoleAdminUUID,
		Status:   "pending",
	}

	// [BARU] Karena di database Deskripsi & PhotoURL boleh kosong (pointer *string)
	// Kita bikin logic pengecekannya dulu
	var deskripsiPtr *string
	if req.Deskripsi != "" {
		deskripsiPtr = &req.Deskripsi
	}

	var photoPtr *string
	if req.PhotoURL != "" {
		photoPtr = &req.PhotoURL
	}

	app := &domain.AdminApplication{
		ID:          uuid.New(),
		UserID:      userID,
		NamaApotek:  req.NamaApotek,
		Alamat:      req.Alamat,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		PhoneNumber: req.PhoneNumber,
		Deskripsi:   deskripsiPtr, // [BARU] Masukkan ke sini
		PhotoURL:    photoPtr,     // [BARU] Masukkan ke sini
		Status:      "PENDING",
	}

	err := s.UserRepo.RegisterAdminTx(user, app)
	if err != nil {
		return err
	}

	data := struct {
		Name       string
		NamaApotek string
	}{
		Name:       req.Name,
		NamaApotek: req.NamaApotek,
	}

	emailBody, err := utils.ParseTemplate("templates/emails/admin_registration.html", data)
	if err != nil {
		log.Println("❌ Error template admin_registration:", err)
		emailBody = "Pendaftaran Admin Apotek Anda berhasil diajukan. Mohon tunggu verifikasi 2x24 jam."
	}

	utils.SendEmail(req.Email, "Pendaftaran Admin Medifinder Diterima", emailBody)

	return nil
}
