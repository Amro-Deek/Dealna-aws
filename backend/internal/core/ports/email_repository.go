package ports

type IEmailService interface {
	//SendEmail(to string, subject string, body string) error
	SendActivationLink(to string, link string, userType string) error
	SendApplicationStatusEmail(to string, status string, comment string) error
}
