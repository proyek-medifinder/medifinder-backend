package service

import (
	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type ResepService struct {
	Repo *repository.ResepRepository
}

func (s *ResepService) Create(userID string, filePath string) error {

	resep := &domain.Resep{
		ID:       uuid.New(),
		UserID:   uuid.MustParse(userID),
		FilePath: filePath,
	}

	return s.Repo.Create(resep)
}

func (s *ResepService) List(limit, offset int) ([]domain.Resep, int, error) {
	return s.Repo.FindAll(limit, offset)
}

func (s *ResepService) UpdateStatus(id string, status string) error {
	return s.Repo.UpdateStatus(uuid.MustParse(id), status)
}
