package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type ApotekService struct {
	Repo *repository.ApotekRepository
}

func (s *ApotekService) Create(adminID string, nama, alamat string, lat, long float64) error {

	existing, _ := s.Repo.FindByAdmin(adminID)
	if existing != nil {
		return errors.New("admin already has an apotek")
	}

	apotek := &domain.Apotek{
		ID:        uuid.New(),
		AdminID:   uuid.MustParse(adminID),
		Nama:      nama,
		Alamat:    alamat,
		Latitude:  lat,
		Longitude: long,
	}

	return s.Repo.Create(apotek)
}

func (s *ApotekService) GetMyApotek(adminID string) (*domain.Apotek, error) {
	return s.Repo.FindByAdmin(adminID)
}

func (s *ApotekService) Update(adminID string, nama, alamat string, lat, long float64) error {

	apotek, err := s.Repo.FindByAdmin(adminID)
	if err != nil {
		return err
	}

	apotek.Nama = nama
	apotek.Alamat = alamat
	apotek.Latitude = lat
	apotek.Longitude = long

	return s.Repo.Update(apotek)
}

func (s *ApotekService) SearchNearby(
	lat, lng, radius float64,
	limit, offset int,
) ([]domain.Apotek, int, error) {

	return s.Repo.FindNearby(lat, lng, radius, limit, offset)
}
