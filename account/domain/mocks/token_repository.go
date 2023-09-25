package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"time"
)

type MockTokenRepository struct {
	mock.Mock
}

// SetRefreshToken is a mock of model.TokenRepository SetRefreshToken
func (m *MockTokenRepository) SetRefreshToken(ctx context.Context, userID, tokenID string, expiresIn time.Duration) error {
	args := m.Called(ctx, userID, tokenID, expiresIn)

	var r0 error
	if args.Get(0) != nil {
		r0 = args.Get(0).(error)
	}

	return r0
}

// DeleteRefreshToken is a mock of model.TokenRepository DeleteRefreshToken
func (m *MockTokenRepository) DeleteRefreshToken(ctx context.Context, userID, prevTokenID string) error {
	args := m.Called(ctx, userID, prevTokenID)

	var r0 error
	if args.Get(0) != nil {
		r0 = args.Get(0).(error)
	}

	return r0
}
