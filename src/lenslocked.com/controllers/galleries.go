package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/eitah/lenslocked/src/lenslocked.com/context"
	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
	"github.com/gorilla/mux"
)

func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		NewView:        views.NewView("bootstrap", "galleries/new"),
		ShowView:       views.NewView("bootstrap", "galleries/show"),
		GalleryService: gs,
	}
}

type Galleries struct {
	NewView        *views.View
	ShowView       *views.View
	GalleryService models.GalleryService
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// GET /galleries/new
func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	g.NewView.Render(w, nil)
}

// GET /galleries/show/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return
	}

	gallery, err := g.GalleryService.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			http.Error(w, "Whoops!Something went wrong.",
				http.StatusInternalServerError)
		}
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, vd)
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

	// This is what the validator code is for, keeping this from being brittle.
	user := context.User(r.Context())
	gallery := models.Gallery{
		UserID: user.ID,
		Title:  form.Title,
	}

	if err := g.GalleryService.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, vd)
		return
	}
	fmt.Fprintln(w, "Gallery is", gallery)
}
