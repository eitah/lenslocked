package main

import (
	"fmt"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Welcome to my great site!</h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Contact page</h1><p>To get in touch with me, reach out to <a href=\"mailto:elijahbit@gmail.com\">elijahbit@gmail.com</a>.</p>")
}

func fourohfourHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>Zomg page not found!</h1>")

}

type Router struct{}

func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/":
		homeHandler(w, r)
	case "/contact":
		contactHandler(w, r)
	default:
		fourohfourHandler(w, r)
	}
}

func main() {
	var router Router
	fmt.Println("Starting the server on :3000")
	http.ListenAndServe(":3000", router)
}
