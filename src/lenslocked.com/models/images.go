package models

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	"google.golang.org/api/iterator"
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
	db  *gorm.DB
	bkt *storage.BucketHandle
}

func NewImageService(db *gorm.DB, bkt *storage.BucketHandle) ImageService {
	return &imageService{
		db:  db,
		bkt: bkt,
	}
}

func (is *imageService) Create(galleryID uint, r io.Reader, filename string) error {
	galleryPath, err := is.mkImageDir(galleryID)
	if err != nil {
		return err
	}
	path := filepath.Join(galleryPath, filename)
	obj := is.bkt.Object(path)
	w := obj.NewWriter(context.Background())
	if _, err := io.Copy(w, r); err != nil {
		panic(err)
	}
	if err := w.Close(); err != nil {
		panic(err)
	}

	return nil
}

func (is *imageService) Delete(i *Image) error {
	ctx := context.Background()
	if err := is.bkt.Object(i.Filename).Delete(ctx); err != nil {
		return err
	}

	return os.Remove(i.RelativePath())
}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	dir := is.imageDir(galleryID)

	ctx := context.Background()
	itr := is.bkt.Objects(ctx, &storage.Query{Prefix: dir})
	var done bool

	var imgs int
	for !done {
		img, err := itr.Next()
		if err == iterator.Done {
			break
		} else {
			if err != nil {
				panic(err)
			}
		}

		// check if file exists before fetching again
		if _, err := os.Stat(img.Name); err == nil {
			// path/to/whatever exists so just exit range loop
			continue
		} else if errors.Is(err, os.ErrNotExist) {
			// path/to/whatever does *not* exist

			obj := is.bkt.Object(img.Name)
			r, err := obj.NewReader(ctx)
			if err != nil {
				panic(err)
			}

			dst, err := os.Create(img.Name)
			if err != nil {
				return nil, err
			}
			defer dst.Close()

			// todo feature to not download images again.
			if _, err = io.Copy(dst, r); err != nil {
				return nil, err
			}
			imgs++
		} else {
			return nil, err
		}
	}

	spew.Printf("all done with downloads: %d", imgs)

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
