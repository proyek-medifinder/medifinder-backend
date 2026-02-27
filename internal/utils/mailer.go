package utils

import (
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

func SendEmail(to, subject, body string) {
	go func() {
		m := gomail.NewMessage()
		m.SetHeader("From", os.Getenv("SMTP_EMAIL"))
		m.SetHeader("To", to)
		m.SetHeader("Subject", subject)
		m.SetBody("text/html", body)

		d := gomail.NewDialer(
			os.Getenv("SMTP_HOST"),
			465,
			os.Getenv("SMTP_EMAIL"),
			os.Getenv("SMTP_PASS"),
		)

		if err := d.DialAndSend(m); err != nil {
			log.Println("Email gagal dikirim:", err)
		}
	}()
}
