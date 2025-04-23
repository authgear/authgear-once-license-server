package smtp

import (
	"gopkg.in/gomail.v2"
)

type EmailOptions struct {
	Sender   string
	Subject  string
	HTMLBody string
	To       string
}

type NewDialerOptions struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

func NewDialer(options NewDialerOptions) *gomail.Dialer {
	return gomail.NewDialer(options.SMTPHost, options.SMTPPort, options.SMTPUsername, options.SMTPPassword)
}

func SendEmail(dialer *gomail.Dialer, options EmailOptions) error {
	m := gomail.NewMessage()

	m.SetHeader("From", options.Sender)

	m.SetHeader("To", options.To)

	m.SetHeader("Subject", options.Subject)

	m.SetBody("text/html", options.HTMLBody)

	err := dialer.DialAndSend(m)
	if err != nil {
		return err
	}

	return nil
}
