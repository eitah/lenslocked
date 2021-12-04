package controllers

import (
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
)

func NewGalleries() *Galleries {
	return &Galleries{
		NewView: views.NewView("bootstrap", "galleries/new"),
	}
}

type Galleries struct {
	NewView *views.View
}
