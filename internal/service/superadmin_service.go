package service

import (
	"errors"
	"os"

	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/repository"
	"github.com/sasaefulanwar/medifinder/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type SuperAdminService struct {
	UserRepo *repository.UserRepository
}

func (s *SuperAdminService) ListAdmin(limit, offset int) ([]domain.User, int, error) {
	return s.UserRepo.FindAdmins(limit, offset)
}

func (s *SuperAdminService) CreateAdmin(name, email, password string) error {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &domain.User{
		ID:       uuid.New(),
		Name:     name,
		Email:    email,
		Password: string(hashed),
		RoleID:   RoleAdminUUID,
	}

	return s.UserRepo.CreateAdmin(user)
}

func (s *SuperAdminService) UpdateAdmin(id, name, email string) error {
	return s.UserRepo.UpdateAdmin(uuid.MustParse(id), name, email)
}

func (s *SuperAdminService) DeleteAdmin(id string) error {
	// 1. TUKER STRING ID MENJADI FORMAT UUID
	adminUUID, err := uuid.Parse(id)
	if err != nil {
		return err // Kalau frontend ngirim ID ngaco, langsung stop di sini
	}

	// 2. KIRIM UUID TADI KE REPOSITORY USER BUAT DIHAPUS (Pake DeleteAdmin!)
	err = s.UserRepo.DeleteAdmin(adminUUID)
	if err != nil {
		return err // Kalau database nolak, error-nya dilempar ke sini biar gak crash
	}

	return nil
}

func (s *SuperAdminService) GetPendingAdmins(limit, offset int) ([]dto.PendingAdminResponse, int, error) {
	return s.UserRepo.FindPendingAdmins(limit, offset)
}

func (s *SuperAdminService) ChangeAdminStatus(adminID string, status string) error {
	id, err := uuid.Parse(adminID)
	if err != nil {
		return errors.New("format ID admin tidak valid")
	}

	// Validasi input status biar nggak diisi sembarangan
	if status != "approved" && status != "suspended" {
		return errors.New("status hanya bisa 'approved' atau 'suspended'")
	}

	return s.UserRepo.UpdateAdminStatus(id, status)
}

// Pastikan package utils juga di-import di atas: "github.com/sasaefulanwar/medifinder/internal/utils"

func (s *SuperAdminService) VerifyAdmin(adminIDStr, superAdminIDStr, action, notes string) error {
	if action != "approved" && action != "rejected" {
		return errors.New("action harus 'approved' atau 'rejected'")
	}

	if action == "rejected" && notes == "" {
		return errors.New("alasan penolakan (notes) wajib diisi")
	}

	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return errors.New("invalid admin ID")
	}

	superAdminID, err := uuid.Parse(superAdminIDStr)
	if err != nil {
		return errors.New("invalid super admin ID")
	}

	adminData, err := s.UserRepo.GetUserProfile(adminID)
	if err != nil {
		return errors.New("admin tidak ditemukan")
	}

	appData, _ := s.UserRepo.GetAdminApplicationByUserID(adminID)
	namaApotek := "Apotek Anda"
	if appData != nil {
		namaApotek = appData.NamaApotek
	}

	// 2. Eksekusi Transaction
	err = s.UserRepo.ProcessAdminVerificationTx(adminID, superAdminID, action, notes)
	if err != nil {
		return err
	}

	// 3. Kirim Email Hasil Verifikasi pake Template
	go func() {
		var subject, body, templatePath string
		baseURL := os.Getenv("APP_URL")
		if baseURL == "" {
			baseURL = "http://localhost:3000"
		}

		if action == "approved" {
			subject = "Selamat! Akun Admin Medifinder Disetujui 🎉"
			templatePath = "templates/emails/admin_approved.html"

			data := struct {
				Name       string
				NamaApotek string
				LoginLink  string
			}{
				Name:       adminData.Name,
				NamaApotek: namaApotek,
				LoginLink:  baseURL + "/login",
			}
			body, _ = utils.ParseTemplate(templatePath, data)
		} else {
			subject = "Update Pendaftaran Admin Medifinder"
			templatePath = "templates/emails/admin_rejected.html"

			data := struct {
				Name       string
				NamaApotek string
				Reason     string
			}{
				Name:       adminData.Name,
				NamaApotek: namaApotek,
				Reason:     notes,
			}
			body, _ = utils.ParseTemplate(templatePath, data)
		}

		// Kalau template gagal di-load, body bakal kosong. Kita kasih fallback.
		if body == "" {
			if action == "approved" {
				body = "Pendaftaran apotek Anda disetujui. Silakan login."
			} else {
				body = "Pendaftaran ditolak: " + notes
			}
		}

		utils.SendEmail(adminData.Email, subject, body)
	}()

	return nil
}
