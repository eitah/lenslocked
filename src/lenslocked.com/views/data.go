package views

import (
	"log"

	"github.com/eitah/lenslocked/src/lenslocked.com/models"
)

const (
	AlertLvlError   = "danger"
	AlertLvlWarning = "warning"
	AlertLvlInfo    = "info"
	AlertLvlSuccess = "success"

	// AlertMessageGeneric is displayed whenever a random error is encountered by our backend
	AlertMessageGeneric = "Something went wrong, Please try again and contact us if the problem persists."
)

// Data is the top level structure that views expect data to come in.
type Data struct {
	Alert *Alert
	Yield interface{}
	User  *models.User
}

// Alert is the Boostrap Alert message template.
type Alert struct {
	Level   string
	Message string
}

// Publicerror is an error that can be exposed publicly.
type PublicError interface {
	error
	Public() string
}

func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

func (d *Data) SetAlert(err error) {
	var msg string
	if pErr, ok := err.(PublicError); ok {
		msg = pErr.Public()
	} else {
		log.Println(err)
		msg = AlertMessageGeneric
	}
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}
