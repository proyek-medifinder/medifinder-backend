package utils

import (
	"bytes"
	"html/template"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

func SendEmail(to, subject, body string) {
	go func() {
		apiKey := os.Getenv("RESEND_API_KEY")

		client := resend.NewClient(apiKey)

		params := &resend.SendEmailRequest{
			From:    "onboarding@resend.dev",
			To:      []string{to},
			Subject: subject,
			Html:    body,
		}

		_, err := client.Emails.Send(params)

		if err != nil {
			log.Println("Email gagal dikirim:", err)
			return
		}

		log.Println("Email berhasil dikirim ke", to)
	}()
}

func ParseTemplate(templateFileName string, data interface{}) (string, error) {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}