package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eitah/lenslocked/src/lenslocked.com/context"
	"github.com/eitah/lenslocked/src/lenslocked.com/email"
	"github.com/eitah/lenslocked/src/lenslocked.com/models"
	"github.com/eitah/lenslocked/src/lenslocked.com/rand"
	"github.com/eitah/lenslocked/src/lenslocked.com/views"
	"github.com/gorilla/mux"
)

func NewUsers(us models.UserService, emailClient email.EmailClient, r *mux.Router) *Users {
	return &Users{
		NewView:      views.NewView("bootstrap", "users/new"),
		LoginView:    views.NewView("bootstrap", "users/login"),
		ForgotPWView: views.NewView("bootstrap", "users/forgot_pw"),
		ResetPWView:  views.NewView("bootstrap", "users/reset_pw"),
		UserService:  us,
		Email:        emailClient,
		r:            r,
	}
}

type Users struct {
	NewView      *views.View
	LoginView    *views.View
	ForgotPWView *views.View
	ResetPWView  *views.View
	UserService  models.UserService
	Email        email.EmailClient
	r            *mux.Router
}

type SignupForm struct {
	Email    string `schema:"email"`
	Name     string `schema:"name"`
	Password string `schema:"password"`
	Age      uint   `schema:"age"`
}

// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	parseURLParams(r, &form) // ensures that if there is form data in query params will be prefilled.
	u.NewView.Render(w, r, form)
}

// GET /forgot
func (u *Users) Forgot(w http.ResponseWriter, r *http.Request) {
	var form ResetPWForm
	parseURLParams(r, &form) // ensures that if there is form data in query params will be prefilled.
	u.ForgotPWView.Render(w, r, form)
}

// GET /reset
func (u *Users) Reset(w http.ResponseWriter, r *http.Request) {
	var form ResetPWForm
	parseURLParams(r, &form) // ensures that if there is form data in query params will be prefilled.
	u.ResetPWView.Render(w, r, form)
}

// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form SignupForm
	vd.Yield = &form // persist data on redirect. note we have to use a pointer here to make sure as values are updated form remembers data too.
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.NewView.Render(w, r, vd)
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
		u.NewView.Render(w, r, vd)
		return
	}

	if err := u.signIn(w, &user); err != nil {
		// we assume its soem short lived data outage and so try to let users just proceed.
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := u.Email.SendWelcomeEmail(); err != nil {
		fmt.Printf("Error sending welcome email: %s\n", err)
		// continuing because this is not a fatal error
	}

	url, err := u.r.Get(IndexGalleries).URL()
	if err != nil {
		vd.AlertError(fmt.Sprintf("Something went wrong: %s", err))
		views.RedirectAlert(w, r, "/", http.StatusFound, *vd.Alert)
		return
	}
	views.RedirectAlert(w, r, url.Path, http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: fmt.Sprintf("Welcome to Lenslocked.com, %s!", user.Name),
	})
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form LoginForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
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
		u.LoginView.Render(w, r, vd)
		return
	}

	if err := u.signIn(w, user); err != nil {
		vd.SetAlert(err)
		u.LoginView.Render(w, r, vd)
		return
	}

	url, err := u.r.Get(IndexGalleries).URL()
	if err != nil {
		vd.AlertError(fmt.Sprintf("Something went wrong: %s", err))
		views.RedirectAlert(w, r, "/", http.StatusFound, *vd.Alert)
		return
	}
	views.RedirectAlert(w, r, url.Path, http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: fmt.Sprintf("Welcome to Lenslocked.com, %s!", user.Name),
	})
}

// Logout is used to delete a user's session cookie and invalidate their
// current remember token, which will sign the current user out.
//
// POST /logout
func (u *Users) Logout(w http.ResponseWriter, r *http.Request) {
	// First expire the users cookie
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	// then we add a new remember token
	user := context.User(r.Context())
	// ignore errors because they are 1) unlikely and 2) we cant recover now that
	// we don't have a valid cookie
	token, _ := rand.RememberToken()
	user.Remember = token
	u.UserService.Update(user)
	// Send the user to the home page
	http.Redirect(w, r, "/", http.StatusFound)
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

type ResetPWForm struct {
	Email    string `schema:"email"`
	Token    string `schema:"token"`
	Password string `schema:"password"`
}

// InitiateReset starts a reset password flow
func (u *Users) InitiateReset(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPWForm
	vd.Yield = &form

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		views.RedirectAlert(w, r, "/reset", http.StatusFound, *vd.Alert)
	}

	token, err := u.UserService.InitiateReset(form.Email)
	if err != nil {
		vd.SetAlert(err)
		u.ForgotPWView.Render(w, r, vd)
		return
	}

	if err := u.Email.SendForgotPasswordEmail(token); err != nil {
		vd.SetAlert(err)
		u.ForgotPWView.Render(w, r, vd)
		return
	}

	vd.Alert = &views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: fmt.Sprintf("Reset password sent to %s! You have 12 hours.", form.Email),
	}
	u.ForgotPWView.Render(w, r, vd)
}

// CompleteReset concludes a reset password flow
func (u *Users) CompleteReset(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form ResetPWForm
	vd.Yield = &form

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		views.RedirectAlert(w, r, fmt.Sprintf("/reset?token=%s", form.Token), http.StatusFound, *vd.Alert)
	}

	user, err := u.UserService.CompleteReset(form.Token, form.Password)
	if err != nil {
		vd.SetAlert(err)
		u.ResetPWView.Render(w, r, vd)
		return
	}
	u.signIn(w, user)
	views.RedirectAlert(w, r, "/galleries", http.StatusFound, views.Alert{
		Level:   views.AlertLvlSuccess,
		Message: fmt.Sprintf("Password changed successfully! Enjoy our site, %s", user.Name),
	})
}
