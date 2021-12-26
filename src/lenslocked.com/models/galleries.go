package models

import (
	"github.com/jinzhu/gorm"
)

type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not null;index"`
	Title  string `gorm:"not_null"`
	// Images []string todo come back to this
}

type GalleryService interface {
	GalleryDB
}

type galleryService struct {
	GalleryDB
}

type galleryValidator struct {
	GalleryDB
}

// todo eli doesn't understand what this does or why it matters
var _ GalleryDB = &galleryGorm{}

type GalleryDB interface {
	Create(gallery *Gallery) error
}

func NewGalleryService(db *gorm.DB) GalleryService {
	gg := &galleryGorm{
		db: db,
	}
	gv := NewGalleryValidator(gg)
	return &galleryService{
		GalleryDB: gv,
	}
}

func NewGalleryValidator(gdb GalleryDB) *galleryValidator {
	return &galleryValidator{
		GalleryDB: gdb,
	}
}

type galleryGorm struct {
	db *gorm.DB
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	return gv.GalleryDB.Create(gallery)
}

// 1. write working gallery service create
// 2. controller and view to render the form
// 3. parse the gallery form and create a gallery
// 4. add validations
func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}
