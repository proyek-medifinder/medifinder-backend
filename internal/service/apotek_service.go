package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type ApotekService struct {
	Repo *repository.ApotekRepository
}

func (s *ApotekService) Create(adminID string, nama, alamat string, lat, long float64, jamBuka, jamTutup string) error {
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
		JamBuka:   jamBuka,
		JamTutup:  jamTutup,
	}

	return s.Repo.Create(apotek)
}

func (s *ApotekService) GetMyApotek(adminID uuid.UUID) (*domain.Apotek, error) {
	return s.Repo.FindByAdmin(adminID)
}

func (s *ApotekService) Update(adminID string, nama, alamat string, lat, long float64, jamBuka, jamTutup, phoneNumber, deskripsi string) error {
	apotek, err := s.Repo.FindByAdmin(adminID)
	if err != nil {
		return err
	}

	apotek.Nama = nama
	apotek.Alamat = alamat
	apotek.Latitude = lat
	apotek.Longitude = long
	apotek.JamBuka = jamBuka
	apotek.JamTutup = jamTutup
	apotek.PhoneNumber = phoneNumber // Tambahan field baru
	apotek.Deskripsi = &deskripsi    // Tambahan field baru (karena di domain pake pointer)

	return s.Repo.Update(apotek)
}

func (s *ApotekService) SearchNearby(
	lat, lng, radius float64,
	limit, offset int,
) ([]domain.Apotek, int, error) {

	// Set Timezone ke WIB (Asia/Jakarta)
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local // Fallback kalau library time nggak nemu zona
	}

	// Ambil jam saat ini dengan format HH:MM:SS
	currentTime := time.Now().In(loc).Format("15:04:05")

	return s.Repo.FindNearby(lat, lng, radius, limit, offset, currentTime)
}
