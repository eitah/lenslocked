package middleware

import (
	"fmt"
	"net/http"

	"github.com/eitah/lenslocked/src/lenslocked.com/context"
	"github.com/eitah/lenslocked/src/lenslocked.com/models"
)

type RequireUser struct {
	models.UserService
}

func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// todo check if a user is logged in. If so, call next(w,r)
		// if not http.Redirect to /login
		fmt.Println("in handler")
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// set the user on the context which uses our custom package to ensure typesafety.
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)

		next(w, r)
	})
}

func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}
