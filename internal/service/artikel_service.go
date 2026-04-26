package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sasaefulanwar/medifinder/internal/domain"
	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/repository"
)

type ArtikelService struct {
	Repo *repository.ArtikelRepository
}

// Response structure bawaan NewsAPI
type NewsAPIResponse struct {
	Articles []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		UrlToImage  string `json:"urlToImage"`
		Url         string `json:"url"`
	} `json:"articles"`
}

// Fungsi buat narik berita dari NewsAPI & disimpen ke Database lu
func (s *ArtikelService) FetchHealthNews() error {
	apiKey := os.Getenv("NEWSAPI_KEY")
	if apiKey == "" {
		return fmt.Errorf("API KEY NewsAPI belum di set")
	}

	// Tembak berita kesehatan spesifik Indonesia
	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=id&category=health&apiKey=%s", apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var newsResp NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		return err
	}

	for _, item := range newsResp.Articles {
		if item.Title == "" || item.UrlToImage == "" {
			continue // Skip berita gajelas yang gaada gambarnya
		}

		// Bikin slug simpel dari judul
		slug := strings.ToLower(strings.ReplaceAll(item.Title, " ", "-"))
		if len(slug) > 200 {
			slug = slug[:200]
		}

		// Masukin sebagai artikel
		artikel := &domain.Artikel{
			ID:       uuid.New(),
			Judul:    item.Title,
			Slug:     slug,
			Konten:   item.Description + " \n\nBaca selengkapnya di: " + item.Url,
			Kategori: "Kesehatan Terkini",
			ImageURL: &item.UrlToImage,
			Status:   "PUBLISHED", // Otomatis dipublish
			Source:   "api_news",
		}

		err := s.Repo.Create(artikel)
		if err != nil {
			log.Println("Gagal nyimpen artikel API:", err)
		}
	}

	log.Println("✅ Berhasil nge-fetch artikel kesehatan dari NewsAPI!")
	return nil
}

func (s *ArtikelService) GetPublishedArticles(page, limit int) ([]domain.Artikel, error) {
	offset := (page - 1) * limit
	return s.Repo.GetPublished(limit, offset)
}

// ================= TAMBAHAN BUAT SUPER ADMIN =================

func (s *ArtikelService) CreateManual(superAdminID string, req dto.CreateArtikelRequest, imageURL string) error {
	adminUUID, err := uuid.Parse(superAdminID)
	if err != nil {
		return err
	}

	// Bikin slug dari judul (contoh: "Cara Tidur Cepat" -> "cara-tidur-cepat")
	slug := strings.ToLower(strings.ReplaceAll(req.Judul, " ", "-"))

	artikel := &domain.Artikel{
		ID:           uuid.New(),
		SuperAdminID: &adminUUID, // Tercatat siapa yang nulis
		Judul:        req.Judul,
		Slug:         slug + "-" + uuid.New().String()[:5], // Tambah random string dikit biar ga gampang bentrok
		Konten:       req.Konten,
		Kategori:     req.Kategori,
		Status:       req.Status,
		Source:       "internal", // Tandain ini buatan asli
	}

	if imageURL != "" {
		artikel.ImageURL = &imageURL
	}

	return s.Repo.Create(artikel)
}

func (s *ArtikelService) DeleteArtikel(id string) error {
	return s.Repo.Delete(id)
}
