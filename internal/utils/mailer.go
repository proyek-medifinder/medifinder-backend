package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// SendEmail mengirim email secara asynchronous via REST API Brevo (Port 443 HTTPS - Anti Blokir Cloud)
func SendEmail(to, subject, body string) {
	apiKey := os.Getenv("BREVO_API_KEY")
	senderEmail := os.Getenv("EMAIL_SENDER")
	if senderEmail == "" {
		senderEmail = "cs.medifinder@gmail.com"
	}

	if apiKey == "" {
		log.Println("⚠️ MAILER ERROR: BREVO_API_KEY belum diset di environment variable!")
		return
	}

	go func() {
		payload := map[string]interface{}{
			"sender": map[string]string{
				"name":  "Medifinder",
				"email": senderEmail,
			},
			"to": []map[string]string{
				{
					"email": to,
				},
			},
			"subject":     subject,
			"htmlContent": body,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Println("Email payload marshal error:", err)
			return
		}

		req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println("Email request creation error:", err)
			return
		}

		req.Header.Set("accept", "application/json")
		req.Header.Set("api-key", apiKey)
		req.Header.Set("content-type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("Email gagal dikirim via Brevo API:", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Println("Email berhasil dikirim ke", to)
		} else {
			log.Printf("Email gagal dikirim via Brevo [%d]: %s\n", resp.StatusCode, string(respBody))
		}
	}()
}

// ParseTemplate mencari lokasi file html secara absolut dari root proyek agar anti-error path
func ParseTemplate(templateFileName string, data interface{}) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan working directory: %v", err)
	}

	finalPath := filepath.Join(currentDir, templateFileName)

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
