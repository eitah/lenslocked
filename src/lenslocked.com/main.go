package main

import (
	"fmt"
	"net/http"

	"github.com/eitah/lenslocked/src/lenslocked.com/controllers"
	"github.com/eitah/lenslocked/src/lenslocked.com/middleware"
	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	"github.com/eitah/lenslocked/src/lenslocked.com/rand"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	fourOhFourView *views.View
)

func fourOhFour(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fourOhFourView.Render(w, r, nil)
}

func main() {
	config := DefaultConfig()
	dbConfig := DefaultPostgresConfig()
	services, err := models.NewServices(dbConfig.Dialect(), dbConfig.ConnectionInfo())
	if err != nil {
		panic(err)
	}

	defer services.Close()
	services.AutoMigrate()

	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User, r)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)
	fourOhFourView = views.NewView("bootstrap", "fourohfour")

	userMW := &middleware.User{
		UserService: services.User,
	}
	requireUserMW := &middleware.RequireUser{}

	// Handle lets you just get a view
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.Handle("/pay-me-money", staticC.PayMeMoney).Methods("GET")
	r.Handle("/login", usersC.LoginView).Methods("GET")

	// Handlefunc calls a method on the controller
	// Normally we only need function calls when pagdes are posts, but here we want business logic for alerts
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	r.Handle("/galleries/new", requireUserMW.Apply(galleriesC.NewView)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMW.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/show/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name(controllers.ShowGallery)
	r.HandleFunc("/galleries", requireUserMW.ApplyFn(galleriesC.Index)).Methods("GET").Name(controllers.IndexGalleries)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMW.ApplyFn(galleriesC.Edit)).Methods("GET").Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMW.ApplyFn(galleriesC.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", requireUserMW.ApplyFn(galleriesC.Delete)).Methods("POST")

	// <form action="/galleries/{{.GalleryID}}/images/{{.Filename}}/delete" method="POST">
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requireUserMW.ApplyFn(galleriesC.ImageDelete)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", requireUserMW.ApplyFn(galleriesC.ImageUpload)).Methods("POST")

	// http.Dir matches the path exactly with how it fetches the static file, so we use stripPrefix to help match the path.
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	r.NotFoundHandler = http.HandlerFunc(fourOhFour)
	fmt.Println("Starting server on http://localhost:3000")

	randString, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}

	// although forbidden to do this, unless server restarts
	// this would reset, so this could be a static token instead
	csrfMW := csrf.Protect(randString, csrf.Secure(config.IsProd()))
	port := fmt.Sprintf(":%d", config.Port)
	http.ListenAndServe(port, csrfMW(userMW.Apply(r)))
}
