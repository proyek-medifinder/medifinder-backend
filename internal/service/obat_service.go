package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type ObatService struct {
	ObatRepo   *repository.ObatRepository
	ApotekRepo *repository.ApotekRepository
}

func (s *ObatService) Create(adminID, nama string, deskripsi *string, stok int, harga float64, imageURL *string) (*domain.Obat, error) {
	apotek, err := s.ApotekRepo.FindByAdmin(uuid.MustParse(adminID))
	if err != nil {
		return nil, errors.New("admin has no apotek")
	}

	obat := &domain.Obat{
		ID:        uuid.New(),
		ApotekID:  apotek.ID,
		Nama:      nama,
		Deskripsi: deskripsi,
		Stok:      stok,
		Harga:     harga,
		ImageURL:  imageURL,
	}

	err = s.ObatRepo.Create(obat)
	if err != nil {
		return nil, err
	}

	return obat, nil
}

func (s *ObatService) GetByApotek(apotekID string) ([]domain.Obat, error) {
	return s.ObatRepo.FindByApotek(apotekID)
}

func (s *ObatService) GetMyObat(adminID string) ([]domain.Obat, error) {
	apotek, err := s.ApotekRepo.FindByAdmin(uuid.MustParse(adminID))
	if err != nil {
		return nil, errors.New("admin has no apotek")
	}

	return s.ObatRepo.FindByApotek(apotek.ID.String())
}

func (s *ObatService) GetByID(obatID string) (*domain.Obat, error) {
	return s.ObatRepo.FindByID(obatID)
}

func (s *ObatService) GetMyObatByID(adminID, obatID string) (*domain.Obat, error) {
	apotek, err := s.ApotekRepo.FindByAdmin(uuid.MustParse(adminID))
	if err != nil {
		return nil, errors.New("admin has no apotek")
	}

	obat, err := s.ObatRepo.FindByID(obatID)
	if err != nil {
		return nil, err
	}

	if obat.ApotekID != apotek.ID {
		return nil, errors.New("not your apotek")
	}

	return obat, nil
}

func (s *ObatService) GetPublicByApotek(apotekID string, name string, limit, offset int) ([]domain.Obat, int, error) {
	obat, err := s.ObatRepo.FindByApotekPaginated(apotekID, name, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.ObatRepo.CountByApotek(apotekID, name)
	if err != nil {
		return nil, 0, err
	}

	return obat, total, nil
}

func (s *ObatService) Update(adminID, obatID, nama string, deskripsi *string, stok int, harga float64, imageURL *string) (*domain.Obat, error) {
	apotek, err := s.ApotekRepo.FindByAdmin(uuid.MustParse(adminID))
	if err != nil {
		return nil, errors.New("admin has no apotek")
	}

	obat, err := s.ObatRepo.FindByID(obatID)
	if err != nil {
		return nil, err
	}

	if obat.ApotekID != apotek.ID {
		return nil, errors.New("not your apotek")
	}

	obat.Nama = nama
	obat.Stok = stok
	obat.Harga = harga

	if deskripsi != nil {
		obat.Deskripsi = deskripsi
	}

	if imageURL != nil && *imageURL != "" {
		obat.ImageURL = imageURL
	}

	err = s.ObatRepo.Update(obat)
	if err != nil {
		return nil, err
	}

	return obat, nil
}

func (s *ObatService) UpdateImage(adminID, obatID, imageURL string) error {
	apotek, err := s.ApotekRepo.FindByAdmin(uuid.MustParse(adminID))
	if err != nil {
		return errors.New("admin has no apotek")
	}

	obat, err := s.ObatRepo.FindByID(obatID)
	if err != nil {
		return err
	}

	if obat.ApotekID != apotek.ID {
		return errors.New("not your apotek")
	}

	return s.ObatRepo.UpdateImage(obatID, imageURL)
}

func (s *ObatService) Delete(adminID, obatID string) error {
	apotek, err := s.ApotekRepo.FindByAdmin(uuid.MustParse(adminID))
	if err != nil {
		return errors.New("admin has no apotek")
	}

	obat, err := s.ObatRepo.FindByID(obatID)
	if err != nil {
		return err
	}

	if obat.ApotekID != apotek.ID {
		return errors.New("not your apotek")
	}

	return s.ObatRepo.Delete(obatID)
}
