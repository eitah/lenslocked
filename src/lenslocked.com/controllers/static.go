package controllers

import "github.com/eitah/lenslocked/src/lenslocked.com/views"

func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "static/home"),
		Contact: views.NewView("bootstrap", "static/contact"),
	}

}

type Static struct {
	Home    *views.View
	Contact *views.View
}
