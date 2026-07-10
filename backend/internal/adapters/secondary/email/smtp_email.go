package email

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/Amro-Deek/Dealna-aws/backend/internal/config"
	"github.com/Amro-Deek/Dealna-aws/backend/internal/core/ports"
)

type SMTPEmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

var _ ports.IEmailService = (*SMTPEmailService)(nil)

func NewSMTPEmailService(cfg config.SMTPConfig) *SMTPEmailService {
	return &SMTPEmailService{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
	}
}

func (s *SMTPEmailService) SendActivationLink(email, link, userType string) error {
	if s.host == "mock" || s.host == "" {
		fmt.Printf("📧 [MOCK EMAIL] To: %s, Link: %s, Type: %s\n", email, link, userType)
		return nil
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	var tpl string
	if userType == "provider" {
		tpl = providerEmailTemplate
	} else {
		tpl = studentEmailTemplate
	}

	// Simple string replacement for the link
	body := strings.ReplaceAll(tpl, "{{link}}", link)

	// HTML Email headers
	headers := "MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"From: " + s.from + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: Activate your Dealna Account\r\n\r\n"

	msg := []byte(headers + body)

	return smtp.SendMail(addr, auth, s.from, []string{email}, msg)
}

func (s *SMTPEmailService) SendApplicationStatusEmail(email, status, comment string) error {
	if s.host == "mock" || s.host == "" {
		fmt.Printf("📧 [MOCK EMAIL] To: %s, Status: %s, Comment: %s\n", email, status, comment)
		return nil
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	subject := "Dealna Provider Application Update"
	var body string
	if status == "APPROVED" {
		body = "<h2>Congratulations!</h2><p>Your application to become a provider on Dealna has been approved.</p><p>You can now log in and start offering your services.</p>"
	} else if status == "REJECTED" {
		body = fmt.Sprintf("<h2>Application Update</h2><p>Unfortunately, your application requires some changes.</p><p><strong>Reason:</strong> %s</p><p>Please log in to the app to update your documents and resubmit.</p>", comment)
	} else {
		return nil // Ignore other statuses
	}

	headers := "MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"From: " + s.from + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n\r\n"

	msg := []byte(headers + body)

	return smtp.SendMail(addr, auth, s.from, []string{email}, msg)
}
