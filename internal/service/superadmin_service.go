package service

import (
	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
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
		RoleID:   2,
	}

	return s.UserRepo.CreateAdmin(user)
}

func (s *SuperAdminService) UpdateAdmin(id, name, email string) error {
	return s.UserRepo.UpdateAdmin(uuid.MustParse(id), name, email)
}

func (s *SuperAdminService) DeleteAdmin(id string) error {
	return s.UserRepo.DeleteAdmin(uuid.MustParse(id))
}
