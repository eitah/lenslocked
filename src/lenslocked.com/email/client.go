package email

import (
	"gopkg.in/mailgun/mailgun-go.v1"
)

type EmailClient struct {
	client           mailgun.Mailgun
	elisEmailAddress string
}

func NewEmailClient(domain, apikey, publickey, elisEmailAddress string) EmailClient {
	return EmailClient{
		client:           mailgun.NewMailgun(domain, apikey, publickey),
		elisEmailAddress: elisEmailAddress,
	}
}

func (m *EmailClient) SendWelcomeEmail() error {
	from := "support@lenslocked.com"
	subject := "Welcome to lenslocked!"
	text := `
	Hi There!
	Welcome to our awesome website! Thanks for joining!

	Enjoy,
	Lenslocked Support`
	to := m.elisEmailAddress // todo only sends email to me because, well, its a free account
	msg := m.client.NewMessage(from, subject, text, to)
	_, _, err := m.client.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
