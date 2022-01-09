package controllers

import (
	"net/http"
	"net/url"

	"github.com/gorilla/schema"
)

func parseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseValues(r.PostForm, dst)
}

func parseURLParams(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	return parseValues(r.Form, dst)
}

func parseValues(f url.Values, dst interface{}) error {
	dec := schema.NewDecoder()
	// IgnoreUnknownKeys tells schema to not panic if the CSRF token is unused.
	dec.IgnoreUnknownKeys(true)
	return dec.Decode(dst, f)
}
