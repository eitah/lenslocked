package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/eitah/lenslocked/src/lenslocked.com/hash"
	"github.com/eitah/lenslocked/src/lenslocked.com/rand"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrNotFound is return when a resource cannot be found in the db
	ErrNotFound modelError = "models: resource not found"
	// ErrIDInvalid is returned when an invalid id is provided to delete
	ErrIDInvalid modelError = "models: ID provided was invalid"
	// ErrPasswordIncorrect is returned when an invalid password is provided
	ErrPasswordIncorrect modelError = "models: incorrect password provided"
	// ErrEmailRequired is returned when an email isn't provided
	ErrEmailRequired modelError = "models: email not provided"
	// ErrEmailInvalid is returned when our email does not get regexed
	ErrEmailInvalid modelError = "models: email invalid according to regex"
	// ErrEmailTaken indicates email has already been claimed
	ErrEmailTaken modelError = "models: email already in use"
	// ErrPasswordTooShort indicates an invalid password length.
	ErrPasswordTooShort modelError = "models: password must be at least 3 characters long"
	// ErrPasswordRequired indicates a password was not provided when creating.
	ErrPasswordRequired modelError = "models: password is required"
	// ErrRememberRequired means a remember token is not present for create or update, suggesting a bug.
	ErrRememberRequired modelError = "models: remember token is required"
	// ErrRememberTooShort means our remember token is somehow invalid
	ErrRememberTooShort modelError = "models: remmember token too short"
	// ErrPWResetTokenInvalid means our password Reset Token is somehow invalid
	ErrPWResetTokenInvalid modelError = "models: password reset token provided is invalid"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	sanitizedString := strings.Replace(string(e), "models: ", "", 1)
	final := strings.Replace(sanitizedString,
		sanitizedString[0:1],
		strings.ToUpper(sanitizedString[0:1]),
		1)
	return final
}

// User represents a db user.
// gorm annotations put constraints on the db
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Age          uint
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

type UserService interface {
	// Authenticate will verify the provided email address and password.
	// If they are correct the user corresponding to the email will be returned.
	// Otherwise you get an error if something goes wrong.
	Authenticate(email, password string) (*User, error)
	// InitiateReset will complete all the model related tasks to start the pw reset
	// process for the user with the provided email address. Once completed, it will
	// return the token, or an error if there is one
	InitiateReset(email string) (string, error)
	// CompleteReset will complete all the model-related tasks to complete the password
	// reset process for the user that the token matches, incuding updating that users
	// password. If the token has expired or if it is invalid for any other reason, the
	// ErrTokenInvalid error will be returned.
	CompleteReset(token, newPw string) (*User, error)
	UserDB
}

type userService struct {
	UserDB
	pepper    string
	pwResetDB pwResetDB
}

type userValidator struct {
	UserDB
	hmac       hash.HMAC
	pepper     string
	emailRegex *regexp.Regexp
}

// userGorm reperesents our DB interaction layer and implements
// the userDB interface fully
type userGorm struct {
	db *gorm.DB
}

type UserDB interface {
	// Methods for querying single users
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)
	ByAge(age uint) (*User, error)

	// Methods for querying multiple users
	InAgeRange(min uint, max uint) ([]*User, error)

	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error
}

func NewUserService(db *gorm.DB, pepper, hmacSecretKey string) UserService {
	ug := &userGorm{
		db: db,
	}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := NewUserValidator(ug, hmac, pepper)
	return &userService{
		UserDB:    uv,
		pepper:    pepper,
		pwResetDB: NewPwResetValidator(&pwResetGorm{db: db}, hmac),
	}
}

func NewUserValidator(udb UserDB, hmac hash.HMAC, pepper string) *userValidator {
	return &userValidator{
		UserDB:     udb,
		hmac:       hmac,
		pepper:     pepper,
		emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	pepperedPWBytes := []byte(password + us.pepper)
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), pepperedPWBytes); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, ErrPasswordIncorrect
		} else {
			return nil, err
		}
	}

	return foundUser, nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return ErrEmailInvalid
	}
	return nil
}

// ByID will look up a user with the provided ID.
// If the user is found we will return a nil error.
// If the user is not found we will return ErrNotFound
// If there is another error we will return it.
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (uv *userValidator) ByEmail(email string) (*User, error) {
	user := User{
		Email: email,
	}

	if err := runUserValFns(&user, uv.normalizeEmail); err != nil {
		return nil, err
	}
	return uv.UserDB.ByEmail(user.Email)
}

func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ug *userGorm) ByAge(age uint) (*User, error) {
	var user User
	db := ug.db.Where("age = ?", age)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// InAgeRange finds users within an age range exclusive
func (ug *userGorm) InAgeRange(min uint, max uint) ([]*User, error) {
	var users []*User
	if err := ug.db.Find(&users, "age > ? AND age < ? ", min, max).Error; err != nil {
		panic(err)
	}
	return users, nil
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	if err := runUserValFns(&user, uv.hmacRemember); err != nil {
		return nil, err
	}

	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

// ByRemember looks up a user with the given remember token and returns that user
// expects user token to already be hashed
func (ug *userGorm) ByRemember(token string) (*User, error) {
	var user User
	db := ug.db.Where("remember_hash = ?", token)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// first handles boilerplate that would otherwise have to be copied everywhere
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		// if pw is unchanged no need to re-hash the password
		return nil
	}
	// the pepper is a const that we merge with the password just to up entropy.
	pepperedPWBytes := []byte(user.Password + uv.pepper)
	// DefaultCost is a const representing computing power needed for working the hash, recognizing that the higher the cost the more expensive the app.
	hashedBytes, err := bcrypt.GenerateFromPassword(pepperedPWBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// heres a typical hashed password
	// $2a$10$KdgNj2kbSgl8PuKi0lrmNua.U5Ax5QgjeFzhy96/X4304XzvuC64u
	// $2a$ - the format of the password hash
	// $10 - the cost of the hash
	// the rest - first half is the salt second half is the hashed salted pw
	user.PasswordHash = string(hashedBytes)
	// it isnt necessary to wipe out the password but we do it so the plantext password is never logged.
	user.Password = ""
	return nil
}

func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uv *userValidator) emailIsAvail(user *User) error {
	existing, err := uv.UserDB.ByEmail(user.Email)
	if err == ErrNotFound {
		// email is available if it's not found
		return nil
	}
	if err != nil {
		return err
	}

	// if is current user, allow it, else fail validation.
	if user.ID != existing.ID {
		return ErrEmailTaken
	}
	return nil
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}
	if len(user.Password) < 3 {
		return ErrPasswordTooShort
	}
	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return ErrRememberRequired
	}

	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}
	bytes, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if bytes < rand.RememberTokenBytes {
		return ErrRememberTooShort
	}
	return nil
}

// Create creates the provided user and backfills data like the id and cretaedat fields
func (uv *userValidator) Create(user *User) error {
	if err := runUserValFns(user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.requireEmail,
		uv.normalizeEmail,
		uv.emailFormat,
		uv.emailIsAvail); err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

func (uv *userValidator) Update(user *User) error {
	if err := runUserValFns(user,
		uv.passwordMinLength,
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.requireEmail,
		uv.normalizeEmail,
		uv.emailFormat,
		uv.emailIsAvail); err != nil {
		return err
	}
	return uv.UserDB.Update(user)
}

func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

func (uv *userValidator) idGreaterThan(id uint) userValFn {
	return userValFn(func(user *User) error {
		if user.ID <= id {
			return ErrIDInvalid
		}
		return nil
	})
}

func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	if err := runUserValFns(&user, uv.idGreaterThan(0)); err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(user).Error
}

type userValFn func(*User) error

func runUserValFns(user *User, fns ...userValFn) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func (us *userService) InitiateReset(email string) (string, error) {
	user, err := us.ByEmail(email)
	if err != nil {
		return "", err
	}
	pwr := pwReset{
		UserID: user.ID,
	}
	if err := us.pwResetDB.Create(&pwr); err != nil {
		return "", err
	}
	return pwr.Token, nil
}

func (us *userService) CompleteReset(token, newPw string) (*User, error) {
	pwr, err := us.pwResetDB.ByToken(token)
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrPWResetTokenInvalid
		}
		return nil, err
	}
	duration := time.Since(pwr.CreatedAt)
	if duration > 12*time.Hour {
		fmt.Printf("Password Reset token for user %d is too old: %d\n", pwr.UserID, duration)
		return nil, ErrPWResetTokenInvalid
	}

	user, err := us.ByID(pwr.UserID)
	if err != nil {
		return nil, err
	}
	user.Password = newPw
	err = us.Update(user)
	if err != nil {
		return nil, err
	}
	us.pwResetDB.Delete(pwr.ID)
	return user, nil
}
