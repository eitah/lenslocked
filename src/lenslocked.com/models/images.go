package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jinzhu/gorm"
)

// Image is used to represent images stored in a gallery.
// Instead references images stored on disk.
type Image struct {
	GalleryID uint
	Filename  string
}

func (i *Image) Path() string {
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return temp.String()
}

func (i *Image) RelativePath() string {
	// convert the gallery id to a string
	galleryID := fmt.Sprintf("%v", i.GalleryID)
	return filepath.ToSlash(filepath.Join("images", "galleries", galleryID, i.Filename))
}

type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error
	Delete(i *Image) error
	ByGalleryID(galleryID uint) ([]Image, error)
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

func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	dir := is.imageDir(galleryID)
	strings, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}

	ret := make([]Image, len(strings))
	// Adds leading "/"" to all image file paths
	for i, imageStr := range strings {
		ret[i] = Image{
			Filename:  filepath.Base(imageStr),
			GalleryID: galleryID,
		}
	}

	return ret, nil
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
