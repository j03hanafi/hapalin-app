package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGet(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		uid, _ := uuid.NewRandom()

		mockResp := &domain.User{
			UID:   uid,
			Email: "bob@bob.com",
			Name:  "Bobby Bobson",
		}

		mockUserRepository := new(mocks.MockUserRepository)
		mockUserRepository.
			On("FindByID", mock.Anything, uid).
			Return(mockResp, nil)

		us := NewUserService(&USConfig{
			UserRepository: mockUserRepository,
		})

		ctx := context.TODO()
		u, err := us.Get(ctx, uid)

		assert.NoError(t, err)
		assert.Equal(t, mockResp, u)
		mockUserRepository.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		uid, _ := uuid.NewRandom()

		mockUserRepository := new(mocks.MockUserRepository)
		mockUserRepository.
			On("FindByID", mock.Anything, uid).
			Return(nil, assert.AnError)

		us := NewUserService(&USConfig{
			UserRepository: mockUserRepository,
		})

		ctx := context.TODO()
		u, err := us.Get(ctx, uid)

		assert.Nil(t, u)
		assert.Error(t, err)
		mockUserRepository.AssertExpectations(t)
	})
}
