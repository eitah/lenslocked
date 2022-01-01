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
	ByGalleryID(galleryID uint) ([]string, error)
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
	galleryPath, err := is.mkImageDir(galleryID)
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

func (is *imageService) ByGalleryID(galleryID uint) ([]string, error) {
	dir := is.imageDir(galleryID)
	strings, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}

	// Adds leading "/"" to all image file paths
	for i := range strings {
		strings[i] = "/" + strings[i]
	}

	return strings, nil
}

func (is *imageService) mkImageDir(galleryID uint) (string, error) {
	dir := is.imageDir(galleryID)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func (is *imageService) imageDir(galleryID uint) string {
	return filepath.Join("images", "galleries", fmt.Sprintf("%v", galleryID))
}
