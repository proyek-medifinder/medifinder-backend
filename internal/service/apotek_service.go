package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type ApotekService struct {
	Repo *repository.ApotekRepository
}

func (s *ApotekService) Create(adminID string, nama, alamat string, lat, long float64, jamBuka, jamTutup string, photo_url *string) error {
	adminUUID := uuid.MustParse(adminID)

	existing, _ := s.Repo.FindByAdmin(adminUUID)
	if existing != nil {
		return errors.New("admin already has an apotek")
	}

	apotek := &domain.Apotek{
		ID:        uuid.New(),
		AdminID:   adminUUID, // Note: Sesuaikan kalau lu pake ApotekID atau apalah di domain
		Nama:      nama,
		Alamat:    alamat,
		Latitude:  lat,
		Longitude: long,
		JamBuka:   &jamBuka,  // Map parameter jamBuka
		JamTutup:  &jamTutup, // Map parameter jamTutup
		PhotoURL:  photo_url, // Map parameter gambar
	}

	if apotek.PhotoURL != nil {
		fmt.Printf("DEBUG SERVICE: PhotoURL domain: %s\n", *apotek.PhotoURL)
	}

	return s.Repo.Create(apotek)
}

func (s *ApotekService) GetByAdmin(adminID string) (*domain.Apotek, error) {
	adminUUID, err := uuid.Parse(adminID)
	if err != nil {
		return nil, errors.New("invalid admin id")
	}

	apotek, err := s.Repo.FindByAdmin(adminUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("apotek tidak ditemukan untuk admin ini")
		}
		return nil, err
	}

	return apotek, nil
}

func (s *ApotekService) Update(adminID string, nama, alamat string, lat, long float64, jamBuka, jamTutup, phoneNumber, deskripsi *string, photo_url *string) error {
	apotek, err := s.Repo.FindByAdmin(uuid.MustParse(adminID)) // Parse dulu!
	if err != nil {
		return err
	}

	apotek.Nama = nama
	apotek.Alamat = alamat
	apotek.Latitude = lat
	apotek.Longitude = long

	apotek.JamBuka = jamBuka
	apotek.JamTutup = jamTutup
	apotek.PhoneNumber = phoneNumber
	apotek.Deskripsi = deskripsi
	apotek.PhotoURL = photo_url

	return s.Repo.Update(apotek)
}

func (s *ApotekService) SearchNearby(
	lat, lng, radius float64,
	limit, offset int,
) ([]domain.Apotek, int, error) {

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.Local
	}

	currentTime := time.Now().In(loc).Format("15:04:05")

	return s.Repo.FindNearby(lat, lng, radius, limit, offset, currentTime)
}

func (s *ApotekService) GetByID(id string) (domain.Apotek, error) {
	apotek, err := s.Repo.GetByID(id)
	if err != nil {
		return apotek, err
	}
	return apotek, nil
}

func (s *ApotekService) UpdateImage(adminID string, photo_url string) error {
	// Cari apotek milik admin ini
	apotek, err := s.Repo.FindByAdmin(uuid.MustParse(adminID))
	if err != nil {
		return errors.New("apotek tidak ditemukan untuk admin ini")
	}

	// Set URL gambar barunya (karena di domain pake *string, kita reference variabelnya)
	apotek.PhotoURL = &photo_url

	// Pake fungsi Update bawaan repo lu yang udah ada
	return s.Repo.Update(apotek)
}

func (s *ApotekService) GetAllForSuperAdmin(page, limit int) ([]domain.Apotek, int, error) {
	offset := (page - 1) * limit
	return s.Repo.FindAllWithCount(limit, offset)
}
