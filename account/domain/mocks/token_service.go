package mocks

import (
	"context"
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
