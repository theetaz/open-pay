package smtp

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// Config holds SMTP connection configuration.
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Sender sends emails via SMTP.
type Sender struct {
	cfg Config
}

// NewSender creates a new SMTP email sender.
func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

// SendEmail sends an HTML email to the given recipient.
func (s *Sender) SendEmail(to, subject, htmlBody string) error {
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	headers := []string{
		fmt.Sprintf("From: %s", s.cfg.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
	}

	msg := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, msg)
}
