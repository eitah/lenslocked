package controllers

import (
	"fmt"
	"net/http"

	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
)

func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView:     views.NewView("bootstrap", "users/new"),
		LoginView:   views.NewView("bootstrap", "users/login"),
		UserService: us,
	}
}

type Users struct {
	NewView     *views.View
	LoginView   *views.View
	UserService *models.UserService
}

type SignupForm struct {
	Email    string `schema:"email"`
	Name     string `schema:"name"`
	Password string `schema:"password"`
	Age      uint   `schema:"age"`
}

// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if err := ParseForm(r, &form); err != nil {
		panic(err)
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Age:      form.Age,
		Password: form.Password,
	}

	if err := u.UserService.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "User is", user)
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var form LoginForm
	if err := ParseForm(r, &form); err != nil {
		panic(err)
	}

	user, err := u.UserService.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			fmt.Fprintln(w, "Invalid Email Address")
		case models.ErrInvalidPassword:
			fmt.Fprintln(w, "Invalid Password")
		default:
			http.Error(w, fmt.Sprintf("unhandled error %s", err.Error()), http.StatusInternalServerError)
		}
		return
	}

	cookie := http.Cookie{
		Name:  "email",
		Value: user.Email,
	}
	http.SetCookie(w, &cookie)

	fmt.Fprintln(w, user)
}

// Five main attack vectors for cookies in jons course
// cookie tampering
// a db leak that lets users make fake cookies
// cross site scripting
// cookie theft via packet sniffing
// cookie theft via physical access to the device with the cookie.

// CookieTest is a dev method to see what our cookies like without needing to muck around in devtools
func (u *Users) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("email")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprintln(w, "Email is:", cookie.Value)
}
