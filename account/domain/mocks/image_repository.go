package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"mime/multipart"
)

// MockImageRepository is a mock of domain.ImageRepository
type MockImageRepository struct {
	mock.Mock
}

// UpdateProfile is a mock representation of domain.ImageRepository UpdateProfile
func (m *MockImageRepository) UpdateProfile(ctx context.Context, objName string, imageFile multipart.File) (string, error) {
	args := m.Called(ctx, objName, imageFile)

	var r0 string
	if args.Get(0) != nil {
		r0 = args.Get(0).(string)
	}

	var r1 error
	if args.Get(1) != nil {
		r1 = args.Get(1).(error)
	}

	return r0, r1
}
