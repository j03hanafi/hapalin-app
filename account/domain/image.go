package domain

import (
	"context"
	"mime/multipart"
)

// ImageRepository defines methods it expects a repository
// it interacts with to implement
type ImageRepository interface {
	UpdateProfile(ctx context.Context, objName string, imageFile multipart.File) (string, error)
}
