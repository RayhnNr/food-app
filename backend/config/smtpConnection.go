package config

import (
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
)

func SendEmail(recipient, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	sender := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	if host == "" || port == "" || sender == "" || password == "" {
		return fmt.Errorf("SMTP_HOST, SMTP_PORT, SMTP_EMAIL, and SMTP_PASSWORD must be configured")
	}
	if strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("email subject must not contain line breaks")
	}

	from, err := mail.ParseAddress(sender)
	if err != nil {
		return fmt.Errorf("invalid SMTP sender address: %w", err)
	}
	to, err := mail.ParseAddress(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	message := fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		to.Address,
		from.Address,
		subject,
		body,
	)
	auth := smtp.PlainAuth("", from.Address, password, host)
	if err := smtp.SendMail(net.JoinHostPort(host, port), auth, from.Address, []string{to.Address}, []byte(message)); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
