package mocks

import (
	"context"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

// Get is a mock of UserService.Get
func (m *MockUserService) Get(ctx context.Context, uid uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, uid)

	var r0 *domain.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*domain.User)
	}

	var r1 error
	if args.Get(1) != nil {
		r1 = args.Error(1)
	}

	return r0, r1
}

// SignUp is a mock of UserService.SignUp
func (m *MockUserService) SignUp(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)

	var r0 error
	if args.Get(0) != nil {
		r0 = args.Error(0)
	}

	return r0
}
