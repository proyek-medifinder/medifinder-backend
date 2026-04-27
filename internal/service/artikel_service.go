package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

func (s *ArtikelService) FetchHealthNews() error {
	apiKey := os.Getenv("NEWSAPI_KEY")
	if apiKey == "" {
		return fmt.Errorf("API KEY NewsAPI belum di set")
	}

	queryParam := url.QueryEscape("kesehatan OR medis OR apotek")

	targetURL := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&language=id&sortBy=publishedAt&apiKey=%s", queryParam, apiKey)

	// 3. Tembak API-nya
	resp, err := http.Get(targetURL)
	if err != nil {
		return fmt.Errorf("request ditolak NewsAPI: %v", err)
	}
	defer resp.Body.Close()

	// 4. Bikin proteksi tambahan biar ketahuan kalau API Key salah / Limit habis
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NewsAPI ngasih error status code: %d", resp.StatusCode)
	}

	var newsResp NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		return fmt.Errorf("gagal parsing JSON dari NewsAPI: %v", err)
	}

	// 5. Looping data buat disimpen ke Database
	for _, item := range newsResp.Articles {
		if item.Title == "" || item.UrlToImage == "" {
			continue // Skip berita gajelas yang gaada gambarnya
		}

		// Bikin slug simpel dari judul
		slug := strings.ToLower(strings.ReplaceAll(item.Title, " ", "-"))
		if len(slug) > 200 {
			slug = slug[:200]
		}

		artikel := &domain.Artikel{
			ID:       uuid.New(),
			Judul:    item.Title,
			Slug:     slug + "-" + uuid.New().String()[:5], // Biar slug bener-bener unik
			Konten:   item.Description + " \n\nBaca selengkapnya di: " + item.Url,
			Kategori: "Kesehatan Terkini",
			ImageURL: &item.UrlToImage,
			Status:   "PUBLISHED",
			Source:   "api_news",
		}

		err := s.Repo.Create(artikel)
		if err != nil {
			log.Println("Gagal nyimpen artikel API:", err)
		}
	}

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

func (s *ArtikelService) GetDetailBySlug(slug string) (*domain.Artikel, error) {
	return s.Repo.FindBySlug(slug)
}

func (s *ArtikelService) UpdateArtikel(id string, req dto.UpdateArtikelRequest, imageURL string) error {
	// 1. Cari data artikel aslinya dulu
	artikel, err := s.Repo.FindByID(id)
	if err != nil {
		return fmt.Errorf("artikel tidak ditemukan")
	}

	// 2. Update field-nya
	artikel.Judul = req.Judul
	// Update slug kalau judulnya diganti
	artikel.Slug = strings.ToLower(strings.ReplaceAll(req.Judul, " ", "-")) + "-" + id[:5]
	artikel.Konten = req.Konten
	artikel.Kategori = req.Kategori
	artikel.Status = req.Status

	// Update gambar cuma kalau ada file baru yang diupload
	if imageURL != "" {
		artikel.ImageURL = &imageURL
	}

	return s.Repo.Update(artikel)
}
