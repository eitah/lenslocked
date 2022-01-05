package middleware

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	finalHandler := http.HandlerFunc(final)
	mux.Handle("/", middlewareOne(middlewareTwo(finalHandler)))

	log.Println("listening on :3000")
	err := http.ListenAndServe(":3000", mux)
	log.Fatal(err)
}

func middlewareOne(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Running middleware one")
		next.ServeHTTP(w, r)
		log.Println("running middlware one again")
	})
}

func middlewareTwo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Running middleware two")
		if r.URL.Path == "/foo" {
			return
		}

		next.ServeHTTP(w, r)
		log.Println("running middlware two again")
	})
}

func final(w http.ResponseWriter, r *http.Request) {
	log.Println("Executing final handler")
	w.Write([]byte("OK"))
}
