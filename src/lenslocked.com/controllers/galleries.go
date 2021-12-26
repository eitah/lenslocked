package controllers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
)

func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		NewView:        views.NewView("bootstrap", "galleries/new"),
		GalleryService: gs,
	}
}

type Galleries struct {
	NewView        *views.View
	GalleryService models.GalleryService
}

type GalleryForm struct {
	UserID uint   `schema:"userID"`
	Title  string `schema:"title"`
}

// GET /galleries/new
func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	g.NewView.Render(w, nil)
}

// post /galleries/new
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm

	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}

	gallery := models.Gallery{
		UserID: form.UserID,
		Title:  form.Title,
	}

	if err := g.GalleryService.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}
}
