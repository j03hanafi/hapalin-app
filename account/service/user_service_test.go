package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
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

func TestSignUp(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		uid, _ := uuid.NewRandom()

		mockUser := &domain.User{
			Email:    "bob@bob.com",
			Password: "atLeast8Chars",
		}

		mockUserRepository := new(mocks.MockUserRepository)
		us := NewUserService(&USConfig{
			UserRepository: mockUserRepository,
		})

		// We can use Run method to modify the user when the Create method is called.
		// We can then chain on a Return method to return no error
		mockUserRepository.
			On("Create", mock.AnythingOfType("context.todoCtx"), mockUser).
			Run(func(args mock.Arguments) {
				userArg := args.Get(1).(*domain.User)
				userArg.UID = uid
			}).
			Return(nil)

		ctx := context.TODO()
		err := us.SignUp(ctx, mockUser)
		assert.NoError(t, err)

		// assert user now has a userID
		assert.Equal(t, uid, mockUser.UID)

		mockUserRepository.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		mockUser := &domain.User{
			Email:    "bob@bob.com",
			Password: "atLeast8Chars",
		}

		mockUserRepository := new(mocks.MockUserRepository)
		us := NewUserService(&USConfig{
			UserRepository: mockUserRepository,
		})

		mockErr := apperrors.NewConflict("email", mockUser.Email)

		mockUserRepository.
			On("Create", mock.AnythingOfType("context.todoCtx"), mockUser).
			Return(mockErr)

		ctx := context.TODO()
		err := us.SignUp(ctx, mockUser)
		assert.EqualError(t, err, mockErr.Error())

		mockUserRepository.AssertExpectations(t)
	})
}
