package models

import (
	"errors"

	"github.com/eitah/lenslocked/src/lenslocked.com/hash"
	"github.com/eitah/lenslocked/src/lenslocked.com/rand"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

const hmacSecretKey = "secret-hmac-key"

var (
	// ErrNotFound is return when a resource cannot be found in the db
	ErrNotFound = errors.New("models: resource not found")
	// ErrInvalidID is returned when an invalid id is provided to delete
	ErrInvalidID = errors.New("models: ID provided was invalid")
	// ErrInvalidPassword is returned when an invalid password is provided
	ErrInvalidPassword = errors.New("models: incorrect password provided")

	// userPWPepper - the pepper value is a secret random string that we will save to our config eventually
	userPWPepper = "georgian-kava-licit-unread"
)

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
	Authenticate(email, password string) (*User, error)
	UserDB
}

type userService struct {
	UserDB
}

type userValidator struct {
	UserDB
	hmac hash.HMAC
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

	//Close a DB Connection
	Close() error

	// Migration Helpers
	AutoMigrate() error
	DestructiveReset() error
}

func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}
	hmac := hash.NewHMAC(hmacSecretKey)
	uv := &userValidator{
		hmac:   hmac,
		UserDB: ug,
	}
	return &userService{
		UserDB: uv,
	}, nil
}

func (us *userService) Authenticate(email, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	pepperedPWBytes := []byte(password + userPWPepper)
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), pepperedPWBytes); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, ErrInvalidPassword
		} else {
			return nil, err
		}
	}

	return foundUser, nil
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &userGorm{
		db: db,
	}, nil
}

func (ug *userGorm) Close() error {
	return ug.db.Close()
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
	rememberHash := uv.hmac.Hash(token)
	return uv.UserDB.ByRemember(rememberHash)
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

// Create creates the provided user and backfills data like the id and cretaedat fields
func (uv *userValidator) Create(user *User) error {
	// the pepper is a const that we merge with the password just to up entropy.
	pepperedPWBytes := []byte(user.Password + userPWPepper)
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

	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}

		user.Remember = token
		user.RememberHash = uv.hmac.Hash(token)
	}
	return uv.UserDB.Create(user)
}

func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

func (uv *userValidator) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = uv.hmac.Hash(user.Remember)
	}
	return uv.UserDB.Update(user)
}

func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

func (uv *userValidator) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	return uv.UserDB.Delete(id)
}

func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	return ug.db.Delete(user).Error
}

// Nonprod feature
//   1) calls drop table if exists method
//   2) rebuild the users table using autoMigrate
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return ug.AutoMigrate()
}

// Automigrate will attempt to auto migrate the users table - its a prod
// safe version of destructivereset
func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}
