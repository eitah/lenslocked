package models

import (
	"github.com/jinzhu/gorm"
)

func NewServices(connectionInfo string) (*Services, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)

	return &Services{
		Gallery: NewGalleryService(db),
		User:    NewUserService(db),
		db:      db,
	}, nil
}

type Services struct {
	Gallery GalleryService
	User    UserService
	db      *gorm.DB
}

func (s *Services) Close() {
	s.db.Close()
}

// Nonprod feature
//   1) calls drop table if exists method
//   2) rebuild the users table using autoMigrate
func (s *Services) DestructiveReset() error {
	if err := s.db.DropTableIfExists(&User{}, &Gallery{}).Error; err != nil {
		return err
	}
	return s.AutoMigrate()
}

// Automigrate will attempt to auto migrate the users table - its a prod
// safe version of destructivereset
func (s *Services) AutoMigrate() error {
	if err := s.db.AutoMigrate(&User{}, &Gallery{}).Error; err != nil {
		return err
	}
	return nil
}
