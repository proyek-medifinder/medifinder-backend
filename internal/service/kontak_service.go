package service

import (
	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type KontakService struct {
	Repo *repository.KontakRepository
}

func (s *KontakService) SubmitMessage(req dto.CreateKontakRequest) error {
	kontak := &domain.Kontak{
		ID:     uuid.New(),
		Nama:   req.Nama,
		Email:  req.Email,
		Subjek: req.Subjek,
		Pesan:  req.Pesan,
		Status: "UNREAD",
	}

	err := s.Repo.Create(kontak)
	if err != nil {
		return err
	}

	// [OPSIONAL] Lu bisa manggil utils.SendEmail di sini buat ngirim notif ke email Super Admin
	// go utils.SendEmail("admin.medifinder@gmail.com", "Pesan Baru: "+req.Subjek, req.Pesan)

	return nil
}

func (s *KontakService) GetMessages(page, limit int) ([]domain.Kontak, error) {
	offset := (page - 1) * limit
	return s.Repo.FindAll(limit, offset)
}

func (s *KontakService) UpdateStatus(id, status string) error {
	return s.Repo.UpdateStatus(id, status)
}
