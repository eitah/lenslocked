package controllers

import (
	"net/http"

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

// GET /galleries/new
func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	if err := g.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}
