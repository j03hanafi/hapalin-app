package repository

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
)

type gcImageRepository struct {
	Storage    *storage.Client
	BucketName string
}

// NewImageRepository is a factory for initializing User Repository
func NewImageRepository(storage *storage.Client, bucketName string) domain.ImageRepository {
	return &gcImageRepository{
		Storage:    storage,
		BucketName: bucketName,
	}
}

// UpdateProfile updates user profile image
func (r gcImageRepository) UpdateProfile(ctx context.Context, objName string, imageFile multipart.File) (string, error) {
	l := logger.Get()

	bucket := r.Storage.Bucket(r.BucketName)

	object := bucket.Object(objName)
	writer := object.NewWriter(ctx)

	// set cache control so profile image will be serve fresh by browser
	// To do this with object handle, we first have to upload, then update
	writer.ObjectAttrs.CacheControl = "Cache-Control:no-cache, max-age=0"

	// multipart.File is an io.Reader, so we can just copy it to the writer
	if _, err := io.Copy(writer, imageFile); err != nil {
		l.Error("Unable to write file to Google Cloud Storage",
			zap.Error(err),
		)
		return "", apperrors.NewInternal()
	}

	if err := writer.Close(); err != nil {
		l.Error("Unable to close Google Cloud Storage writer",
			zap.Error(err),
		)
		return "", fmt.Errorf("unable to close google cloud storage writer: %w", err)
	}

	imageURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", r.BucketName, objName)

	return imageURL, nil
}
