package email

import (
	"fmt"
	"net/url"

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

const resetHTMLTmpl = `Hi there!<br/>
<br/>
It appears that you have requested a password reset. If this was you, please follow the link below <br/>
<a href="%s">%s</a><br/>
<br/>
If you are asked for a token, please use the following value:<br/>
<br/>
%s<br/>
<br/>
If you didn't request a password reset you can safely ignore this email and your account will n <br/>

Best,<br/>
LensLocked Support<br/>
`

// const resetBaseURL = "https://itah-lenslocked.herokuapp.com/reset"
const resetBaseURL = "localhost:3000/reset"

func (m *EmailClient) SendForgotPasswordEmail(token string) error {
	from := "support@lenslocked.com"
	subject := "Password reset request recieved for Lenslocked.com"
	to := m.elisEmailAddress // todo only sends email to me because, well, its a free account
	text := `
	Hi There!

	It appears that you have requested a password reset. If this was you please follow the link below = append(It appears that you have requested a password reset. If this was you please follow the link below,

	%s

	If you are asked for a token please ue the following value

	%s

	If you didnt request a password reset you can safely ignore this email and your account will not be affected.

	Best,
	Lenslocked Support`

	v := url.Values{}
	v.Set("token", token)
	resetURL := resetBaseURL + "?" + v.Encode()
	resetText := fmt.Sprintf(text, resetURL, token)
	resetHTML := fmt.Sprintf(resetHTMLTmpl, resetURL, resetURL, token)
	message := mailgun.NewMessage(from, subject, resetText, to)
	message.SetHtml(resetHTML)
	_, _, err := m.client.Send(message)
	return err
}
