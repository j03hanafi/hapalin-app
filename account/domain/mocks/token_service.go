package mocks

import (
	"context"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/stretchr/testify/mock"
)

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) NewPairFromUser(ctx context.Context, user *domain.User, prevTokenID string) (*domain.TokenPair, error) {
	ret := m.Called(ctx, user, prevTokenID)

	var r0 *domain.TokenPair
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*domain.TokenPair)
	}

	var r1 error
	if ret.Get(1) != nil {
		r1 = ret.Get(1).(error)
	}

	return r0, r1
}

func (m *MockTokenService) ValidateIDToken(tokenString string) (*domain.User, error) {
	ret := m.Called(tokenString)

	var r0 *domain.User
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*domain.User)
	}

	var r1 error
	if ret.Get(1) != nil {
		r1 = ret.Get(1).(error)
	}

	return r0, r1
}

func (m *MockTokenService) ValidateRefreshToken(refreshTokenString string) (*domain.RefreshToken, error) {
	ret := m.Called(refreshTokenString)

	var r0 *domain.RefreshToken
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(*domain.RefreshToken)
	}

	var r1 error
	if ret.Get(1) != nil {
		r1 = ret.Get(1).(error)
	}

	return r0, r1
}

func (m *MockTokenService) SignOut(ctx context.Context, uid uuid.UUID) error {
	ret := m.Called(ctx, uid)

	var r0 error
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(error)
	}

	return r0

}
