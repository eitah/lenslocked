package main

import (
	"fmt"
	"net/http"

	"github.com/eitah/lenslocked/src/lenslocked.com/controllers"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
	"github.com/gorilla/mux"
)

var (
	fourOhFourView *views.View
)

func fourOhFour(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	must(fourOhFourView.Render(w, nil))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers()
	galleriesC := controllers.NewGalleries()

	fourOhFourView = views.NewView("bootstrap", "fourohfour")

	r := mux.NewRouter()
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.Handle("/pay-me-money", staticC.PayMeMoney).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/galleries/new", galleriesC.New).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(fourOhFour)
	fmt.Println("Server starting on :3000...")
	http.ListenAndServe(":3000", r)
}
