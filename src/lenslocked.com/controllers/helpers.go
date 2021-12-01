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
	return dec.Decode(dst, r.PostForm)
}
