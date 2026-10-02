package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
)

// SendEmail mengirim email secara asynchronous (goroutine) menggunakan SMTP Gmail
func SendEmail(to, subject, body string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPass := os.Getenv("SMTP_PASS") // Mengambil dari SMTP_PASS di .env

	if smtpPort == "" {
		smtpPort = "465"
	}

	if smtpHost == "" || smtpEmail == "" || smtpPass == "" {
		log.Println("⚠️ MAILER ERROR: Konfigurasi SMTP belum lengkap di environment variable!")
		return
	}

	go func() {
		addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
		auth := smtp.PlainAuth("", smtpEmail, smtpPass, smtpHost)

		headers := make(map[string]string)
		headers["From"] = fmt.Sprintf("Medifinder <%s>", smtpEmail)
		headers["To"] = to
		headers["Subject"] = subject
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"

		var msg bytes.Buffer
		for k, v := range headers {
			msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		msg.WriteString("\r\n" + body)

		// Jika menggunakan port 465, wajib gunakan TLS Dial langsung
		if smtpPort == "465" {
			tlsConfig := &tls.Config{
				ServerName: smtpHost,
			}

			conn, err := tls.Dial("tcp", addr, tlsConfig)
			if err != nil {
				log.Println("Email gagal konek TLS:", err)
				return
			}
			defer conn.Close()

			client, err := smtp.NewClient(conn, smtpHost)
			if err != nil {
				log.Println("Email gagal inisialisasi client:", err)
				return
			}
			defer client.Quit()

			if err = client.Auth(auth); err != nil {
				log.Println("Email gagal auth SMTP:", err)
				return
			}

			if err = client.Mail(smtpEmail); err != nil {
				log.Println("Email gagal set sender:", err)
				return
			}

			if err = client.Rcpt(to); err != nil {
				log.Println("Email gagal set recipient:", err)
				return
			}

			w, err := client.Data()
			if err != nil {
				log.Println("Email gagal create data writer:", err)
				return
			}

			if _, err = w.Write(msg.Bytes()); err != nil {
				log.Println("Email gagal write body:", err)
				return
			}

			if err = w.Close(); err != nil {
				log.Println("Email gagal close writer:", err)
				return
			}
		} else {
			// Fallback jika menggunakan port 587
			err := smtp.SendMail(addr, auth, smtpEmail, []string{to}, msg.Bytes())
			if err != nil {
				log.Println("Email gagal dikirim via SMTP:", err)
				return
			}
		}

		log.Println("Email berhasil dikirim ke", to)
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
