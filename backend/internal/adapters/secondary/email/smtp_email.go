package email

import (
	"fmt"
	"net/smtp"

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

func NewSMTPEmailService() *SMTPEmailService {
	return &SMTPEmailService{
		host:     "smtp.gmail.com",
		port:     "587",
		username: "amrobasheer242@gmail.com",
		password: "ydvn uwjc xbqj yvhi",
		from:     "amrobasheer242@gmail.com",
	}
}

func (s *SMTPEmailService) SendActivationLink(email, token string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	link := fmt.Sprintf("http://localhost:8080/api/v1/auth/student/activate?token=%s", token)

	body := fmt.Sprintf(`Activate your Dealna account:

%s
`, link)

	msg := []byte("From: " + s.from + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: Activate Dealna\r\n\r\n" +
		body)

	return smtp.SendMail(addr, auth, s.from, []string{email}, msg)
}
