package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

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

type User struct {
	gorm.Model
	Name string
	// these annotations put constraints on the db
	Email        string `gorm:"not null;unique_index"`
	Age          uint
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
}

func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &UserService{
		db: db,
	}, nil
}

type UserService struct {
	db *gorm.DB
}

func (us *UserService) Close() error {
	return us.db.Close()
}

// ByID will look up a user with the provided ID.
// If the user is found we will return a nil error.
// If the user is not found we will return ErrNotFound
// If there is another error we will return it.
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (us *UserService) ByAge(age uint) (*User, error) {
	var user User
	db := us.db.Where("age = ?", age)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// InAgeRange finds users within an age range exclusive
func (us *UserService) InAgeRange(min uint, max uint) ([]*User, error) {
	var users []*User
	if err := us.db.Find(&users, "age > ? AND age < ? ", min, max).Error; err != nil {
		panic(err)
	}
	return users, nil
}

// first handles boilerplate that would otherwise have to be copied everywhere
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

func (us *UserService) Create(user *User) error {
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
	return us.db.Create(user).Error
}

func (us *UserService) Authenticate(email, password string) (*User, error) {
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

func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{Model: gorm.Model{ID: id}}
	return us.db.Delete(user).Error
}

// Nonprod feature
//   1) calls drop table if exists method
//   2) rebuild the users table using autoMigrate
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

// Automigrate will attempt to auto migrate the users table - its a prod
// safe version of destructivereset
func (us UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}
