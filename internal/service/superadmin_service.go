package service

import (
	"errors"
	"log"
	"os"
	"strings"

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
	adminUUID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("format ID admin tidak valid")
	}

	admin, err := s.UserRepo.DeleteAdmin(adminUUID)
	if err != nil {
		return err
	}

	// Notify only after the deletion has committed successfully.
	body, err := utils.ParseTemplate("templates/emails/admin_deleted.html", struct{ Name string }{admin.Name})
	if err != nil {
		log.Printf("Admin deletion email template failed: %v", err)
	} else {
		utils.SendEmail(admin.Email, "Akun admin apotek dihapus - MediFinder", body)
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
	status = domain.NormalizeAccountStatus(status)
	if status != "active" && status != "suspended" {
		return errors.New("status hanya bisa 'active' atau 'suspended'")
	}

	admin, changed, err := s.UserRepo.UpdateAdminStatus(id, status)
	if err != nil {
		return err
	}
	if changed {
		loginLink := ""
		if baseURL := strings.TrimRight(os.Getenv("APP_URL"), "/"); baseURL != "" {
			loginLink = baseURL + "/admin/login"
		}
		body, err := utils.ParseTemplate("templates/emails/admin_status_changed.html", struct {
			Name      string
			Enabled   bool
			LoginLink string
		}{admin.Name, status == "active", loginLink})
		if err != nil {
			log.Printf("Admin status email template failed: %v", err)
		} else {
			utils.SendEmail(admin.Email, "Perubahan status akun admin - MediFinder", body)
		}
	}
	return nil
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
		var renderErr error
		baseURL := os.Getenv("APP_URL")
		if baseURL == "" {
			baseURL = "http://localhost:3000"
		}

		if action == "approved" {
			subject = "Pengajuan apotek disetujui - MediFinder"
			templatePath = "templates/emails/admin_approved.html"

			data := struct {
				Name       string
				NamaApotek string
				LoginLink  string
			}{
				Name:       adminData.Name,
				NamaApotek: namaApotek,
				LoginLink:  strings.TrimRight(baseURL, "/") + "/admin/login",
			}
			body, renderErr = utils.ParseTemplate(templatePath, data)
		} else {
			subject = "Hasil peninjauan pengajuan apotek - MediFinder"
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
			body, renderErr = utils.ParseTemplate(templatePath, data)
		}

		if renderErr != nil {
			log.Printf("Admin verification email template failed: %v", renderErr)
			return
		}

		utils.SendEmail(adminData.Email, subject, body)
	}()

	return nil
}
