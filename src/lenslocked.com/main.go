package main

import (
	"fmt"
	"net/http"

	"github.com/eitah/lenslocked/src/lenslocked.com/controllers"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
	"github.com/gorilla/mux"
)

var (
	faqView        *views.View
	fourOhFourView *views.View
	payMeMoneyView *views.View
)

func faq(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	must(faqView.Render(w, nil))

}

func payMeMoney(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	must(payMeMoneyView.Render(w, nil))
}

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

	faqView = views.NewView("bootstrap", "views/faq.gohtml")
	fourOhFourView = views.NewView("bootstrap", "views/fourohfour.gohtml")
	payMeMoneyView = views.NewView("bootstrap-nonav", "views/pay-me-money.gohtml")

	r := mux.NewRouter()
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.HandleFunc("/faq", faq).Methods("GET")
	r.HandleFunc("/pay-me-money", payMeMoney).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")

	r.NotFoundHandler = http.HandlerFunc(fourOhFour)
	fmt.Println("Server starting on :3000...")
	http.ListenAndServe(":3000", r)
}
