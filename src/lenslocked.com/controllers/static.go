package controllers

import "github.com/eitah/lenslocked/src/lenslocked.com/views"

func NewStatic() *Static {
	return &Static{
		Home:       views.NewView("bootstrap", "static/home"),
		Contact:    views.NewView("bootstrap", "static/contact"),
		Faq:        views.NewView("bootstrap", "static/faq"),
		PayMeMoney: views.NewView("bootstrap-nonav", "static/pay-me-money"),
	}

}

type Static struct {
	Home       *views.View
	Contact    *views.View
	Faq        *views.View
	PayMeMoney *views.View
}
