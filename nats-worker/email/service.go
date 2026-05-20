package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// Service defines the interface for sending emails.
type Service interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// Config holds SMTP server connection settings.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type smtpService struct {
	cfg Config
}

// NewSMTPService creates a new email service using the SMTP protocol.
func NewSMTPService(cfg Config) Service {
	return &smtpService{cfg: cfg}
}

// SendEmail implements the Service interface for SMTP.
func (s *smtpService) SendEmail(ctx context.Context, to, subject, body string) error {
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	// Build the email message with proper RFC 822 headers
	header := []string{
		fmt.Sprintf("From: %s", s.cfg.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"",
		body,
	}

	message := strings.Join(header, "\r\n")

	// Send the email
	if err := smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}
