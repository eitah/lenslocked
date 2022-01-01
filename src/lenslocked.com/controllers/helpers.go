package controllers

import (
	"net/http"

	"github.com/gorilla/schema"
)

func ParseForm(r *http.Request, dst interface{}) error {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}
	dec := schema.NewDecoder()
	// IgnoreUnknownKeys tells schema to not panic if the CSRF token is unused.
	dec.IgnoreUnknownKeys(true)
	return dec.Decode(dst, r.PostForm)
}
