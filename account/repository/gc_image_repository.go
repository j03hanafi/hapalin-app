package repository

import (
	"cloud.google.com/go/storage"
	"github.com/j03hanafi/hapalin-app/account/domain"
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
