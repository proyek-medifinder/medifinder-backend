package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sasaefulanwar/medifinder/internal/dto"
	"github.com/sasaefulanwar/medifinder/internal/repository"
	openai "github.com/sashabaranov/go-openai"
)

type ChatbotService struct {
	ObatRepo *repository.ObatRepository
}

type groqAIOutput struct {
	Reply        string `json:"reply"`
	MedicineName string `json:"medicine_name"`
}

// Daftar obat keras / berbahaya yang TIDAK BOLEH direkomendasikan bebas
var restrictedMedications = []string{
	"amoxicillin", "amoksisilin", "ciprofloxacin", "antibiotik", "antibiotic",
	"tramadol", "alprazolam", "xanax", "diazepam", "dumolid", "clonazepam",
	"codeine", "morfin", "morphine", "fentanyl", "misoprostol", "cytotec",
	"dexamethasone", "prednisone", "metformin", "amlodipine", "captopril",
}

// Cek apakah obat masuk dalam obat keras
func isRestrictedMedicine(medName string) bool {
	lower := strings.ToLower(medName)
	for _, restricted := range restrictedMedications {
		if strings.Contains(lower, restricted) {
			return true
		}
	}
	return false
}

// Helper cek apakah apotek saat ini sedang buka
func isApotekOpen(jamBuka, jamTutup *string) bool {
	if jamBuka == nil || jamTutup == nil || *jamBuka == "" || *jamTutup == "" {
		return true // Anggap buka jika tidak diatur (24 jam)
	}
	now := time.Now().Format("15:04:05")
	buka := *jamBuka
	tutup := *jamTutup
	if buka <= tutup {
		return now >= buka && now <= tutup
	}
	// Untuk jam operasional lewat tengah malam (misal 13:00 - 02:00)
	return now >= buka || now <= tutup
}
func formatHour(timeStr *string) string {
	if timeStr == nil || *timeStr == "" {
		return "-"
	}
	if len(*timeStr) >= 5 {
		return (*timeStr)[:5] // Ambil format "08:00" dari "08:00:00"
	}
	return *timeStr
}

func (s *ChatbotService) Consult(req dto.ChatRequest) (*dto.ChatResponse, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GROQ_API_KEY is not configured")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1"
	client := openai.NewClientWithConfig(config)

	// Gabungkan informasi usia ke pesan pengguna jika ada
	userPrompt := req.Message
	if req.AgeGroup != "" {
		userPrompt = fmt.Sprintf("[Kategori Usia Pasien: %s (kira-kira %d tahun)]\nKeluhan Pasien: %s", req.AgeGroup, req.Age, req.Message)
	}

	systemPrompt := `Kamu adalah asisten apoteker MediFinder yang beretika, profesional, dan fokus khusus pada kesehatan.

ATURAN KEAMANAN & BATASAN TOPIK (GUARDRAILS):

1. BATASAN DOMAIN (HANYA KESEHATAN & OBAT):
   - Kamu HANYA boleh menjawab pertanyaan seputar: keluhan kesehatan, gejala penyakit ringan, rekomendasi obat bebas (OTC), cara pakai obat, dan informasi apotek.
   - JIKA pengguna bertanya di luar topik kesehatan (misal: teknologi, coding, politik, gosip, tugas sekolah/matematika, resep makanan, cerita dongeng, percintaan, dll):
     TOLAK SECARA RAMAH & TEGAS!
     Contoh respon penolakan: "Maaf, saya adalah asisten apoteker virtual MediFinder yang khusus membantu seputar keluhan kesehatan dan rekomendasi obat. Apakah ada keluhan kesehatan atau pertanyaan obat yang bisa saya bantu? 🩺"
     Isi "medicine_name": ""
   - Jika pengguna sekadar menyapa ramah (halo, hai, selamat pagi, terima kasih), balaslah dengan ramah dan tanyakan keluhan kesehatannya.

2. KONDISI GAWAT DARURAT (RED FLAGS):
   Jika keluhan berupa: nyeri dada kiri menjalar, sesak napas berat, muntah darah, stroke/kelumpuhan mendadak, kejang, pendarahan hebat, atau indikasi bunuh diri/keracunan:
   - JANGAN berikan rekomendasi obat apapun!
   - Berikan respon peringatan darurat: "🚨 Gejala ini berpotensi darurat medis! Segera kunjungi IGD Rumah Sakit terdekat atau hubungi ambulans (119)."
   - Isi "medicine_name": ""

3. LARANGAN OBAT KERAS & ANTIBIOTIK:
   - HANYA rekomendasikan obat bebas (lingkaran hijau) dan bebas terbatas (lingkaran biru) seperti: Paracetamol, Antasida, Cetirizine, Oralit, Obat Batuk Hitam/OBH, Salep gatal hidrokortison ringan, Vitamin.
   - DILARANG merekomendasikan: Antibiotik (misal Amoxicillin), Narkotika, Psikotropika, Obat penenang/tidur, atau Obat Keras (lingkaran merah bertanda 'K').
   - Jika pengguna memaksa meminta obat keras/antibiotik, tolak dengan ramah dan jelaskan bahwa obat tersebut wajib resep dokter.

4. KATEGORI USIA PASIEN:
   - Balita (< 2 tahun) -> Sirup tetes (Drops)
   - Anak (2-12 tahun) -> Sirup anak (bukan tablet dewasa)
   - Dewasa (12-59 tahun) -> Tablet/Kaplet dewasa
   - Lansia (≥ 60 tahun) -> Dosis ramah lansia

FORMAT OUTPUT WAJIB JSON MURNI:
{
  "reply": "Teks jawaban untuk pengguna...",
  "medicine_name": "NamaSatuObat" (atau kosongkan "" jika bukan tentang obat/di luar konteks/kondisi darurat)
}`

	ctx := context.Background()
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "openai/gpt-oss-120b",
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("tidak ada respon dari AI")
	}

	// Parsing JSON dari AI Groq
	var aiResult groqAIOutput
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &aiResult); err != nil {
		return nil, err
	}

	// 🛡️ GUARDRAIL FILTER: Cek apakah nama obat masuk daftar obat keras/terlarang
	if isRestrictedMedicine(aiResult.MedicineName) {
		aiResult.MedicineName = "" // Batalkan pencarian ke database apotek
		aiResult.Reply = "⚠️ Obat yang disebutkan merupakan golongan **Obat Keras / Perlu Resep**. Demi keselamatan kesehatan Anda, obat tersebut memerlukan pemeriksaan langsung dan resep resmi dari dokter. Silakan konsultasikan dengan dokter atau klinik terdekat."
	}

	result := &dto.ChatResponse{
		Reply:        aiResult.Reply,
		MedicineName: aiResult.MedicineName,
		Availability: "NO_MEDICINE_NEEDED",
		IsNearby:     false,
		Pharmacies:   []dto.ChatPharmacyResponse{},
	}

	// Jika ada obat yang direkomendasikan dan user mengirimkan koordinat
	if aiResult.MedicineName != "" && (req.Latitude != 0 || req.Longitude != 0) {
		pharmacies, err := s.ObatRepo.FindPharmaciesWithStock(req.Latitude, req.Longitude, aiResult.MedicineName)
		if err != nil {
			println("❌ Error DB FindPharmaciesWithStock:", err.Error())
		}

		if err == nil && len(pharmacies) > 0 {
			var mapped []dto.ChatPharmacyResponse
			for _, p := range pharmacies {
				mapped = append(mapped, dto.ChatPharmacyResponse{
					ApotekID:  p.ApotekID,
					Nama:      p.Nama,
					Alamat:    p.Alamat,
					Latitude:  p.Latitude,
					Longitude: p.Longitude,
					Distance:  p.Distance,
					Harga:     p.Harga,
					Stok:      p.Stok,
					JamBuka:   formatHour(p.JamBuka),
					JamTutup:  formatHour(p.JamTutup),
					IsOpen:    isApotekOpen(p.JamBuka, p.JamTutup),
				})
			}

			result.Pharmacies = mapped

			const maxNearbyRadius = 10.0 // Batas apotek sekitar (10 km)
			if pharmacies[0].Distance <= maxNearbyRadius {
				result.Availability = "NEARBY"
				result.IsNearby = true
			} else {
				result.Availability = "OUTSIDE_RADIUS"
				result.IsNearby = false
			}
		} else {
			result.Availability = "EMPTY"
			// Tambahkan narasi halus jika obat tidak ditemukan di apotek sekitar
			result.Reply += fmt.Sprintf(
				"\n\nSayang sekali, saat ini apotek di sekitarmu belum ada yang menyediakan %s 😔\n\nKamu bisa mencoba mencarinya melalui fitur Katalog Produk lalu masukkan nama obat yang dicari, siapa tahu ada apotek mitra kami di wilayah lain yang menyediakannya!",
				aiResult.MedicineName,
			)
		}

	}

	return result, nil
}
