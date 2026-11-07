package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/sasaefulanwar/medifinder/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.UserRepo.FindByEmail(strings.TrimSpace(email))
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find reset account: %w", err)
	}

	token, err := utils.GenerateResetToken()
	if err != nil {
		return err
	}
	hash, err := utils.HashResetToken(token)
	if err != nil {
		return err
	}
	link, err := emailLink("/reset-password", url.Values{"token": {token}})
	if err != nil {
		return err
	}
	body, err := utils.ParseTemplate("templates/emails/reset_password.html", struct {
		Name      string
		ResetLink string
	}{user.Name, link})
	if err != nil {
		return err
	}

	issued, err := s.UserRepo.IssuePasswordReset(user.ID, hash)
	if err != nil || !issued {
		return err
	}
	if err := utils.SendEmailNow(user.Email, "Atur ulang kata sandi akun MediFinder", body); err != nil {
		if revokeErr := s.UserRepo.RevokePasswordReset(user.ID, hash); revokeErr != nil {
			return fmt.Errorf("email failed; token cleanup failed: %v: %w", revokeErr, err)
		}
		return err
	}
	return nil
}

func (s *AuthService) ResetPassword(token, newPassword string) error {
	invalid := errors.New("Tautan reset tidak valid, sudah digunakan, atau kedaluwarsa. Silakan minta tautan baru.")
	hash, err := utils.HashResetToken(token)
	if err != nil {
		return invalid
	}
	if len([]rune(newPassword)) < 6 || len([]byte(newPassword)) > 72 {
		return errors.New("Password minimal 6 karakter dan maksimal 72 byte.")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Gagal memproses password baru.")
	}
	email, err := s.UserRepo.ConsumePasswordReset(hash, string(passwordHash))
	if errors.Is(err, sql.ErrNoRows) {
		return invalid
	}
	if err != nil {
		return errors.New("Password belum dapat diperbarui. Silakan coba lagi.")
	}
	s.sendPasswordChangedEmail(email)
	return nil
}
