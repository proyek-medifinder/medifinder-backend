package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
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
	return s.UserRepo.DeleteAdmin(uuid.MustParse(id))
}

func (s *SuperAdminService) GetPendingAdmins(limit, offset int) ([]domain.User, int, error) {
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

	// Ambil email admin buat dikirimin notifikasi nanti
	adminData, err := s.UserRepo.GetUserProfile(adminID)
	if err != nil {
		return errors.New("admin tidak ditemukan")
	}

	// Eksekusi Transaction
	err = s.UserRepo.ProcessAdminVerificationTx(adminID, superAdminID, action, notes)
	if err != nil {
		return err
	}

	// Kirim Email Hasil Verifikasi
	var subject, body string
	if action == "approved" {
		subject = "Selamat! Akun Admin Medifinder Disetujui"
		body = "Aplikasi apotek Anda telah disetujui. Anda sekarang bisa login dan mulai mengelola stok obat."
	} else {
		subject = "Mohon Maaf, Pendaftaran Admin Ditolak"
		body = "Pendaftaran apotek Anda ditolak dengan alasan: " + notes + ". Silakan daftar kembali dengan data yang valid."
	}

	go utils.SendEmail(adminData.Email, subject, body)

	return nil
}
