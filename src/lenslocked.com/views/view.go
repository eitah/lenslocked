package views

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
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
	t, err := template.ParseFiles(files...)
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

func (v *View) Render(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	switch data.(type) {
	case Data:
		// noop - no data wrapping needed.
	default:
		data = Data{
			Yield: data,
		}
	}

	// we are using a buffer because writing any data to response writer
	// results in a 200 status and we can undo the write.
	var buf bytes.Buffer

	if err := v.Template.ExecuteTemplate(&buf, v.Layout, data); err != nil {
		http.Error(w, "Something went wrong. If this error persists, please email support@lenslocked.com", http.StatusInternalServerError)
		return
	}

	// if we get here we know the render succeeded cleanly.
	io.Copy(w, &buf)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, nil)
}
