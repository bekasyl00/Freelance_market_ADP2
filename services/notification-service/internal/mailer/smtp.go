package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"notification_service/internal/models"
)

// Config holds SMTP server connection settings.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type smtpMailer struct {
	cfg Config
}

// NewSMTP creates a new Mailer backed by the SMTP protocol.
func NewSMTP(cfg Config) models.Mailer {
	return &smtpMailer{cfg: cfg}
}

// Send implements the models.Mailer interface.
func (m *smtpMailer) Send(ctx context.Context, email models.Email) error {
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	header := []string{
		fmt.Sprintf("From: %s", m.cfg.From),
		fmt.Sprintf("To: %s", email.To),
		fmt.Sprintf("Subject: %s", email.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"",
		email.Body,
	}

	message := strings.Join(header, "\r\n")

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{email.To}, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}
