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
	"strings"
	"time"
)

// SendEmail preserves asynchronous sending for account notifications.
func SendEmail(to, subject, body string) {
	go func() {
		if err := SendEmailNow(to, subject, body); err != nil {
			log.Printf("Email delivery failed: %v", err)
		}
	}()
}

// SendEmailNow reports whether Brevo accepted the message, not inbox delivery.
func SendEmailNow(to, subject, body string) error {
	return SendEmailNowIdempotent(to, subject, body, "")
}

// Payment retries reuse a stable UUID so Brevo can suppress duplicate requests.
func SendEmailNowIdempotent(to, subject, body, idempotencyKey string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("BREVO_API_KEY is not configured")
	}
	sender := os.Getenv("EMAIL_SENDER")
	if sender == "" {
		sender = "cs.medifinder@gmail.com"
	}
	payload := map[string]interface{}{
		"sender":  map[string]string{"name": "MediFinder", "email": sender},
		"to":      []map[string]string{{"email": to}},
		"subject": subject, "htmlContent": body,
	}
	if idempotencyKey != "" {
		payload["headers"] = map[string]string{"idempotencyKey": idempotencyKey}
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)
	req.Header.Set("content-type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Brevo request failed: %w", err)
	}
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var detail struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(responseBody, &detail) == nil && detail.Message != "" {
			if idempotencyKey != "" && detail.Code == "duplicate_parameter" {
				return nil
			}
			message := strings.ReplaceAll(detail.Code+": "+detail.Message, apiKey, "[REDACTED]")
			if len(message) > 300 {
				message = message[:300]
			}
			return fmt.Errorf("Brevo rejected email (HTTP %d): %s", resp.StatusCode, message)
		}
		return fmt.Errorf("Brevo rejected email (HTTP %d)", resp.StatusCode)
	}
	log.Println("Brevo accepted email for delivery")
	return nil
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

	// Every transactional email shares the same shell and inline email styles.
	layoutPath := filepath.Join(filepath.Dir(finalPath), "layout.html")
	t, err := template.New("email").Funcs(template.FuncMap{
		"currentYear": func() int { return time.Now().Year() },
	}).ParseFiles(layoutPath, finalPath)
	if err != nil {
		return "", fmt.Errorf("gagal parse file template: %v", err)
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "email", data); err != nil {
		return "", fmt.Errorf("gagal execute data ke template: %v", err)
	}

	return buf.String(), nil
}
