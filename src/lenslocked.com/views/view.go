package views

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/eitah/lenslocked/src/lenslocked.com/context"
	"github.com/gorilla/csrf"
)

var (
	LayoutDir   string = "views/layouts/"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

func NewView(layout string, files ...string) *View {
	addTemplateDir(files)
	addTemplateExt(files)
	files = append(files,
		layoutFiles()...,
	)

	t, err := template.New("").Funcs(template.FuncMap{
		// the functions here can be used the same way if, eq, not etc can be used
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("eli the gorilla/csrf lib didnt correctly replace this stub with a valid function so I am blowing up")
		},
		"pathEscape": func(s string) string {
			return url.PathEscape(s)
		},
	}).ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Layout:   layout,
		Template: t,
	}
}

// addTemplateDir takes in a slice of strings
// representing file Dirs for templates, and it
// prepends the templatedir directory to each string
// in the slice
func addTemplateDir(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// addTemplateExt takes ina s lice of strings representing
// file paths and it appends the template ext for each string in the slice

// eg the input {"home"} would result in the output
// {"home.gohtml"} if templateext =  ".gohtml"
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}

func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}
	return files
}

type View struct {
	Template *template.Template
	Layout   string
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	switch d := data.(type) {
	case Data:
		// We need to do this so we can access the data in a var with the type Data.
		vd = d
	default:
		// If the data IS NOT of the type Data, we create one and set the data to
		// the Yield field like before.
		vd = Data{
			Yield: data,
		}
	}

	if alert := getAlert(r); alert != nil {
		vd.Alert = alert
		clearAlert(w)
	}

	vd.User = context.User(r.Context())
	// we are using a buffer because writing any data to response writer
	// results in a 200 status and we can undo the write.
	var buf bytes.Buffer

	//Decorate templates with CSRFFunc
	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})

	if err := tpl.ExecuteTemplate(&buf, v.Layout, vd); err != nil {
		http.Error(w, "Something went wrong. If this error persists, please email support@lenslocked.com", http.StatusInternalServerError)
		return
	}

	// if we get here we know the render succeeded cleanly.
	io.Copy(w, &buf)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}
