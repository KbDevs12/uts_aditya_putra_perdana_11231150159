package config

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

func SendOTPEmail(toEmail, code string) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	username := strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	password := os.Getenv("SMTP_PASSWORD")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if host == "" || port == "" || username == "" || password == "" {
		log.Printf("MVP OTP for %s: %s (SMTP not configured)", toEmail, code)
		return nil
	}
	if from == "" {
		from = username
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	auth := smtp.PlainAuth("", username, password, host)
	subject := "Kode OTP Kantongin"
	body := fmt.Sprintf("Kode OTP Kantongin kamu adalah %s. Kode berlaku 10 menit. Abaikan email ini jika bukan kamu.", code)
	message := "From: " + from + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body

	return smtp.SendMail(host+":"+port, auth, from, []string{toEmail}, []byte(message))
}
