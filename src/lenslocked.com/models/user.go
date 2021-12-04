package models

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Name string
	// these annotations put constraints on the db
	Email string `gorm:"not null;unique_index"`
	// todo make password here, first draft will not have it
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

var (
	// ErrNotFound is return when a resource cannot be found in the db
	ErrNotFound = errors.New("models: resource not found")
)

// ByID will look up a user with the provided ID.
// If the user is found we will return a nil error.
// If the user is not found we will return ErrNotFound
// If there is another error we will return it.
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	err := us.db.Where("id = ?", id).First(&user).Error
	switch err {
	case nil:
		return &user, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("ByID %v", err)
	}
}

func (us *UserService) Create(user *User) error {
	return us.db.Create(user).Error
}

// Nonprod feature
//   1) calls drop table if exists method
//   2) rebuild the users table using autoMigrate
func (us *UserService) DestructiveReset() {
	us.db.DropTableIfExists(&User{})
	us.db.AutoMigrate(&User{})
}
