package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/eitah/lenslocked/src/lenslocked.com/context"
	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
	"github.com/gorilla/mux"
)

const (
	ShowGallery    = "show_gallery"
	IndexGalleries = "index_gallery"
	EditGallery    = "edit_gallery"
)

func NewGalleries(gs models.GalleryService, r *mux.Router) *Galleries {
	return &Galleries{
		NewView:        views.NewView("bootstrap", "galleries/new"),
		ShowView:       views.NewView("bootstrap", "galleries/show"),
		EditView:       views.NewView("bootstrap", "galleries/edit"),
		IndexView:      views.NewView("bootstrap", "galleries/index"),
		GalleryService: gs,
		r:              r,
	}
}

type Galleries struct {
	NewView        *views.View
	ShowView       *views.View
	EditView       *views.View
	IndexView      *views.View
	GalleryService models.GalleryService
	r              *mux.Router
}

type GalleryForm struct {
	Title string `schema:"title"`
}

// GET /galleries/new
func (g *Galleries) New(w http.ResponseWriter, r *http.Request) {
	g.NewView.Render(w, r, nil)
}

// GET /galleries/show/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		// galleryByID handles the error so we just need to return here
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, r, vd)
}

// post /galleries/new
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm

	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.NewView.Render(w, r, vd)
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
		g.NewView.Render(w, r, vd)
		return
	}

	// someday ill understand why redirect to edit view and not index view
	url, err := g.r.Get(EditGallery).URL("id", strconv.Itoa(int(gallery.ID)))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

// GET /galleries
func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	galleries, err := g.GalleryService.ByUserID(uint(user.ID))
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	var vd views.Data
	vd.Yield = galleries
	g.IndexView.Render(w, r, vd)
}

// GET /galleries/:id/edit
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You do not have permission to edit this gallery", http.StatusForbidden)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.EditView.Render(w, r, vd)

}

// POST /galleries/:id/update
func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		// why he uses 404 her and 403 above not sure but w/e
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	var form GalleryForm
	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
	}

	gallery.Title = form.Title

	if err = g.GalleryService.Update(gallery); err != nil {
		vd.SetAlert(err)
	} else {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "Gallery Updated Suceessfully!",
		}
	}

	url, err := g.r.Get(IndexGalleries).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// POST /galleries/:id/delete
func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You do not have permission to edit this gallery", http.StatusForbidden)
		return
	}

	var vd views.Data
	if err := g.GalleryService.Delete(gallery.ID); err != nil {
		vd.SetAlert(err)
		vd.Yield = gallery
		g.EditView.Render(w, r, vd)
		return
	}

	url, err := g.r.Get(IndexGalleries).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}

	gallery, err := g.GalleryService.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			http.Error(w, "Whoops! Something went wrong.",
				http.StatusInternalServerError)
		}
		return nil, err
	}

	return gallery, nil
}

const maxMultipartMem = 1 << 20 // 1 mb

// POST /galleries/:id/images
func (g *Galleries) ImageUpload(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "Gallery not found", http.StatusForbidden)
		return
	}

	// todo bug where the image wont show on the page immediately after upload.

	var vd views.Data
	vd.Yield = gallery
	if err := r.ParseMultipartForm(maxMultipartMem); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	galleryPath := filepath.Join("images", "galleries", fmt.Sprintf("%v", gallery.ID))
	if err := os.MkdirAll(galleryPath, 0755); err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	files := r.MultipartForm.File["images"]
	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		defer file.Close()

		dst, err := os.Create(filepath.Join(galleryPath, f.Filename))
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
	}

	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: "Images Successfully uploaded",
	}
	g.EditView.Render(w, r, vd)
}
