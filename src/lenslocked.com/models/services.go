package models

import (
	"github.com/jinzhu/gorm"
)

type Services struct {
	Gallery GalleryService
	User    UserService
	Image   ImageService
	db      *gorm.DB
}

// named function for declaring service configs
type ServicesConfig func(*Services) error

// for every provided config, iterate pointing to the existing services object.
func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range cfgs {
		// run the function passing in a services pointer
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func WithGorm(dialect, connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.db.LogMode(mode)
		return nil
	}
}

func WithUser(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.db, pepper, hmacKey)
		return nil
	}
}

func WithGallery() ServicesConfig {
	return func(s *Services) error {
		s.Gallery = NewGalleryService(s.db)
		return nil
	}
}

func WithImage() ServicesConfig {
	return func(s *Services) error {
		s.Image = NewImageService(s.db)
		return nil
	}
}

func (s *Services) Close() {
	s.db.Close()
}

// Nonprod feature
//   1) calls drop table if exists method
//   2) rebuild the users table using autoMigrate
func (s *Services) DestructiveReset() error {
	if err := s.db.DropTableIfExists(&User{}, &Gallery{}, &pwReset{}).Error; err != nil {
		return err
	}
	return s.AutoMigrate()
}

// Automigrate will attempt to auto migrate the users table - its a prod
// safe version of destructivereset
func (s *Services) AutoMigrate() error {
	if err := s.db.AutoMigrate(&User{}, &Gallery{}, &pwReset{}).Error; err != nil {
		return err
	}
	return nil
}
