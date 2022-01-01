package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jinzhu/gorm"
)

type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error
	// ByGalleryID(galleryID uint) ([]string, error)
}

type imageService struct {
	db *gorm.DB
}

func NewImageService(db *gorm.DB) ImageService {
	return &imageService{
		db: db,
	}
}

func (is *imageService) Create(galleryID uint, r io.Reader, filename string) error {
	galleryPath, err := is.mkImagePath(galleryID)
	if err != nil {
		return err
	}

	dst, err := os.Create(filepath.Join(galleryPath, filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, r); err != nil {
		return err
	}

	return nil
}

func (is *imageService) mkImagePath(galleryID uint) (string, error) {
	galleryPath := filepath.Join("images", "galleries", fmt.Sprintf("%v", galleryID))
	if err := os.MkdirAll(galleryPath, 0755); err != nil {
		return "", err
	}
	return galleryPath, nil
}
