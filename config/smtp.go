package config

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

func appDisplayName(app string) string {
	switch strings.ToLower(strings.TrimSpace(app)) {
	case "kantongin":
		return "Kantongin"
	case "ecommerce", "e-commerce":
		return "Fragrance App"
	default:
		return "Kantongin"
	}
}

func SendOTPEmail(toEmail, code, app string) error {
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

	brand := appDisplayName(app)
	auth := smtp.PlainAuth("", username, password, host)
	subject := fmt.Sprintf("Kode OTP %s", brand)
	body := fmt.Sprintf("Kode OTP %s kamu adalah %s. Kode berlaku 10 menit. Abaikan email ini jika bukan kamu.", brand, code)
	message := "From: " + from + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body

	return smtp.SendMail(host+":"+port, auth, from, []string{toEmail}, []byte(message))
}
