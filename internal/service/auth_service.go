package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
	"github.com/sasaefulanwar/medifinder/internal/utils"
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

	fmt.Println("EMAIL INPUT:", email)
	fmt.Println("PASSWORD INPUT:", password)

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		fmt.Println("USER NOT FOUND")
		return "", errors.New("invalid email/password")
	}

	fmt.Println("PASSWORD HASH DB:", user.Password)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		fmt.Println("PASSWORD NOT MATCH")
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
