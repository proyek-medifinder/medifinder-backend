package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/resend/resend-go/v2"
)

// SendEmail mengirim email secara asynchronous (goroutine)
func SendEmail(to, subject, body string) {
	// Ambil API Key di luar goroutine untuk memastikan env ter-load dengan benar
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Println("⚠️ MAILER ERROR: RESEND_API_KEY tidak ditemukan di environment variable!")
		return
	}

	// Tentukan asal email secara dinamis (bisa diset lewat env nanti kalau udah punya domain custom)
	fromEmail := os.Getenv("EMAIL_FROM")
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev" // Default resend dev
	}

	go func() {
		client := resend.NewClient(apiKey)

		params := &resend.SendEmailRequest{
			From:    fromEmail,
			To:      []string{to},
			Subject: subject,
			Html:    body,
		}

		_, err := client.Emails.Send(params)
		if err != nil {
			log.Printf("❌ MAILER ERROR [Ke: %s]: Gagal mengirim email. Detail: %v\n", to, err)
			return
		}

		log.Println("📧 MAILER SUCCESS: Email berhasil mendarat ke", to)
	}()
}

// ParseTemplate mencari lokasi file html secara absolut dari root proyek agar anti-error path
func ParseTemplate(templateFileName string, data interface{}) (string, error) {
	// Ambil absolute path dari directory tempat aplikasi berjalan saat ini
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan working directory: %v", err)
	}

	// Gabungkan base path dengan templateFileName (misal: templateFileName isinya "templates/emails/reset_password.html")
	// Kode ini bikin pencarian file aman meskipun dipanggil dari folder handler atau service
	finalPath := filepath.Join(currentDir, templateFileName)

	// Cek dulu filenya beneran ada apa kagak sebelum diparse
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file template tidak ditemukan di jalur: %s", finalPath)
	}

	t, err := template.ParseFiles(finalPath)
	if err != nil {
		return "", fmt.Errorf("gagal parse file template: %v", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("gagal execute data ke template: %v", err)
	}

	return buf.String(), nil
}
