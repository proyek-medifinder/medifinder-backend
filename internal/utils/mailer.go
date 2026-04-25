package utils

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendEmail(to, subject, body string) {
	go func() {
		m := gomail.NewMessage()
		m.SetHeader("From", os.Getenv("SMTP_EMAIL"))
		m.SetHeader("To", to)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", body)

		portStr := os.Getenv("SMTP_PORT")
		port, _ := strconv.Atoi(portStr)
		if port == 0 {
			port = 465
		}

		d := gomail.NewDialer(
			os.Getenv("SMTP_HOST"),
			port,
			os.Getenv("SMTP_EMAIL"),
			os.Getenv("SMTP_PASS"),
		)

		if err := d.DialAndSend(m); err != nil {
			log.Println("Email gagal dikirim:", err)
		}
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
