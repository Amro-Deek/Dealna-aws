package ports

type IEmailService interface {
	//SendEmail(to string, subject string, body string) error
	SendActivationLink(to string, token string) error
}