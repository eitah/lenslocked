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

var (
	ErrUserIDRequired modelError = "models: UserID is required on this gallery"
	ErrTitleRequired  modelError = "models: Title is required on this gallery"
)

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
	return &galleryService{
		GalleryDB: &galleryValidator{
			GalleryDB: &galleryGorm{
				db: db,
			}},
	}
}

type galleryGorm struct {
	db *gorm.DB
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	if err := runGalleryValFns(gallery, []galleryValFn{
		gv.hasValidUserId,
		gv.hasTitle,
	}...); err != nil {
		return err
	}
	return gv.GalleryDB.Create(gallery)
}

func (gv *galleryValidator) hasValidUserId(gallery *Gallery) error {
	if gallery.UserID == 0 {
		return ErrUserIDRequired
	}
	return nil
}

func (gv *galleryValidator) hasTitle(gallery *Gallery) error {
	if gallery.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

// 1. write working gallery service create
// 2. controller and view to render the form
// 3. parse the gallery form and create a gallery
// 4. add validations
func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

type galleryValFn func(*Gallery) error

func runGalleryValFns(gallery *Gallery, fns ...galleryValFn) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}
