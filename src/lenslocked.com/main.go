package main

import (
	"fmt"
	"net/http"

	"github.com/eitah/lenslocked/src/lenslocked.com/controllers"
	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
	"github.com/gorilla/mux"

	_ "github.com/jinzhu/gorm/dialects/postgres"
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

const (
	host   = "localhost"
	port   = 5432
	user   = "eitah"
	dbname = "lenslocked_dev" // this is the dev db
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)
	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer us.Close()
	us.AutoMigrate()

	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(us)
	galleriesC := controllers.NewGalleries()

	fourOhFourView = views.NewView("bootstrap", "fourohfour")

	r := mux.NewRouter()
	// Handle lets you just get a view
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.Handle("/pay-me-money", staticC.PayMeMoney).Methods("GET")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.Handle("/galleries/new", galleriesC.NewView).Methods("GET")

	// Handlefunc calls a method on the controller
	// Normally we only need function calls when pagdes are posts, but here we want business logic for alerts
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(fourOhFour)
	fmt.Println("Server starting on :3000...")
	http.ListenAndServe(":3000", r)
}
