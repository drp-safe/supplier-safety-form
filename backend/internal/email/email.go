package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/obeyler/supplier-safety-form/internal/models"
)

const smtpDialTimeout = 10 * time.Second

type EmailService struct {
	config *models.EmailConfig
}

func NewEmailService(config *models.EmailConfig) *EmailService {
	return &EmailService{config: config}
}

// SendEmail sends a plain HTML email using the configured SMTP server.
func (s *EmailService) SendEmail(to, subject, body string) error {
	message := s.buildMessage(to, subject, body)

	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	switch s.config.SMTPEncryption {
	case "ssl":
		return s.sendWithSSL(addr, auth, to, message)
	case "none":
		return s.sendPlain(addr, auth, to, message)
	default:
		return s.sendWithSTARTTLS(addr, auth, to, message)
	}
}

func (s *EmailService) buildMessage(to, subject, body string) []byte {
	return []byte(fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n",
		s.config.FromName, s.config.FromAddress, to, subject, body))
}

// sendWithSTARTTLS sends email using STARTTLS (most common for port 587).
func (s *EmailService) sendWithSTARTTLS(addr string, auth smtp.Auth, to string, message []byte) error {
	conn, err := net.DialTimeout("tcp", addr, smtpDialTimeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	host, _, _ := net.SplitHostPort(addr)
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()
	if err = client.StartTLS(&tls.Config{ServerName: host}); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return sendMessage(client, s.config.FromAddress, to, message)
}

// sendWithSSL sends email over an implicit TLS connection (port 465).
func (s *EmailService) sendWithSSL(addr string, auth smtp.Auth, to string, message []byte) error {
	host, _, _ := net.SplitHostPort(addr)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: smtpDialTimeout}, "tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return fmt.Errorf("failed to connect over TLS: %w", err)
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return sendMessage(client, s.config.FromAddress, to, message)
}

// sendPlain sends email with no transport encryption (local/dev only).
func (s *EmailService) sendPlain(addr string, auth smtp.Auth, to string, message []byte) error {
	return smtp.SendMail(addr, auth, s.config.FromAddress, []string{to}, message)
}

func sendMessage(client *smtp.Client, from, to string, message []byte) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	if _, err := w.Write(message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return client.Quit()
}
