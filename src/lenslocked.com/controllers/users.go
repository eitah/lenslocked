package controllers

import (
	"fmt"
	"net/http"

	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	"github.com/eitah/lenslocked/src/lenslocked.com/rand"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
)

func NewUsers(us models.UserService) *Users {
	return &Users{
		NewView:     views.NewView("bootstrap", "users/new"),
		LoginView:   views.NewView("bootstrap", "users/login"),
		UserService: us,
	}
}

type Users struct {
	NewView     *views.View
	LoginView   *views.View
	UserService models.UserService
}

type SignupForm struct {
	Email    string `schema:"email"`
	Name     string `schema:"name"`
	Password string `schema:"password"`
	Age      uint   `schema:"age"`
}

// POST /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	u.NewView.Render(w, nil)
}

// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form SignupForm
	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, vd)
		return
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Age:      form.Age,
		Password: form.Password,
	}

	if err := u.UserService.Create(&user); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, vd)
		return
	}

	if err := u.signIn(w, &user); err != nil {
		// we assume its soem short lived data outage and so try to let users just proceed.
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/cookietest", http.StatusNotFound)
	fmt.Fprintln(w, "User is", user)
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form LoginForm
	if err := ParseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, vd)
		return
	}

	user, err := u.UserService.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("No user exists with that email address")
		default:
			vd.SetAlert(err)
		}
		u.LoginView.Render(w, vd)
		return
	}

	if err := u.signIn(w, user); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, vd)
		return
	}

	http.Redirect(w, r, "/cookietest", http.StatusFound)
}

func (u *Users) signIn(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
		if err := u.UserService.Update(user); err != nil {
			return err
		}

		cookie := http.Cookie{
			Name:     "remember_token",
			Value:    user.Remember,
			HttpOnly: true, // tells the cookie that it is not available to scripts.
		}
		http.SetCookie(w, &cookie)
		return nil
	}
	return nil
}

// Five main attack vectors for cookies in jons course
// cookie tampering
// a db leak that lets users make fake cookies
// cross site scripting
// cookie theft via packet sniffing
// cookie theft via physical access to the device with the cookie.

// CookieTest is a dev method to see what our cookies like without needing to muck around in devtools
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintln(w, "<header><meta http-equiv=\"refresh\" content=\"2;url=/login\" /></header><body>Please log in, redirecting to '/login' in 2... 1...</body>")
		return
	}
	fmt.Fprintln(w, "remember me token is:", cookie.Value)
}
