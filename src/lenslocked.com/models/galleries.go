package models

import (
	"github.com/jinzhu/gorm"
)

type Gallery struct {
	gorm.Model
	UserID uint     `gorm:"not null;index"`
	Title  string   `gorm:"not_null"`
	Images []string `gorm:"-"`
}

var (
	ErrUserIDRequired    modelError = "models: UserID is required on this gallery"
	ErrGalleryIdRequired modelError = "models: GalleryID is required"
	ErrTitleRequired     modelError = "models: Title is required on this gallery"
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
	ByID(id uint) (*Gallery, error)
	ByUserID(id uint) ([]*Gallery, error)
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(id uint) error
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

func (gv *galleryValidator) hasValidId(gallery *Gallery) error {
	if gallery.ID == 0 {
		return ErrGalleryIdRequired
	}
	return nil
}

func (gv *galleryValidator) ByID(id uint) (*Gallery, error) {
	gallery := Gallery{
		Model: gorm.Model{
			ID: id,
		},
	}

	if err := runGalleryValFns(&gallery, gv.hasValidId); err != nil {
		return nil, err
	}

	return gv.GalleryDB.ByID(id)
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)
	if err != nil {
		return nil, err
	}
	return &gallery, nil
}

func (gg *galleryGorm) ByUserID(userID uint) ([]*Gallery, error) {
	var galleries []*Gallery
	db := gg.db.Where("user_id = ?", userID)
	if err := db.Find(&galleries).Error; err != nil {
		return nil, err
	}
	return galleries, nil
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

// 1. write working gallery service create
// 2. controller and view to render the form
// 3. parse the gallery form and create a gallery
// 4. add validations
func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gv *galleryValidator) Update(gallery *Gallery) error {
	if err := runGalleryValFns(gallery,
		gv.hasValidUserId,
		gv.hasTitle,
		gv.hasValidId); err != nil {
		return err
	}

	return gv.GalleryDB.Update(gallery)
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gv *galleryValidator) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}

	if err := runGalleryValFns(&gallery, gv.hasValidId); err != nil {
		return err
	}

	return gv.GalleryDB.Delete(id)
}

func (gg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	return gg.db.Delete(&gallery).Error
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
