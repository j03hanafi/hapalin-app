package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/domain/fixture"
	"github.com/j03hanafi/hapalin-app/account/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGet(t *testing.T) {
	// Setup
	t.Parallel()
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

func TestSetProfileImage(t *testing.T) {
	mockUserRepository := new(mocks.MockUserRepository)
	mockImageRepository := new(mocks.MockImageRepository)

	us := NewUserService(&USConfig{
		UserRepository:  mockUserRepository,
		ImageRepository: mockImageRepository,
	})

	t.Run("Successful new image", func(t *testing.T) {
		t.Parallel()

		uid, _ := uuid.NewRandom()

		// mockUser does not have imageURL
		mockUser := &domain.User{
			UID:     uid,
			Email:   "bob@bob.com",
			Website: "https://bob.com",
			Name:    "Bobby Bob",
		}

		findByIDArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			uid,
		}
		mockUserRepository.
			On("FindByID", findByIDArgs...).
			Return(mockUser, nil)

		multipartImageFixture := fixture.NewMultipartImage("image.png", "image/png")
		defer multipartImageFixture.Close()
		imageFileHeader := multipartImageFixture.GetFormFile()
		imageFile, _ := imageFileHeader.Open()

		updateProfileArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mock.AnythingOfType("string"),
			imageFile,
		}

		imageURL := "http://imageurl.com/jdfkj34kljl"

		mockImageRepository.
			On("UpdateProfile", updateProfileArgs...).
			Return(imageURL, nil)

		updateImageArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mockUser.UID,
			imageURL,
		}

		mockUpdatedUser := &domain.User{
			UID:      uid,
			Email:    "newbob@bob.com",
			Website:  "https://bob.com",
			Name:     "New Bob",
			ImageURL: imageURL,
		}

		mockUserRepository.
			On("UpdateImage", updateImageArgs...).
			Return(mockUpdatedUser, nil)

		ctx := context.Background()

		updatedUser, err := us.SetProfileImage(ctx, mockUser.UID, imageFileHeader)

		assert.NoError(t, err)
		assert.Equal(t, mockUpdatedUser, updatedUser)

		mockUserRepository.AssertCalled(t, "FindByID", findByIDArgs...)
		mockImageRepository.AssertCalled(t, "UpdateProfile", updateProfileArgs...)
		mockUserRepository.AssertCalled(t, "UpdateImage", updateImageArgs...)
	})

	t.Run("Successful update image", func(t *testing.T) {
		t.Parallel()

		uid, _ := uuid.NewRandom()
		imageURL := "http://imageurl.com/jdfkj34kljl"

		// has imageURL
		mockUser := &domain.User{
			UID:      uid,
			Email:    "bob@bob.com",
			Website:  "https://bob.com",
			Name:     "Bobby Bob",
			ImageURL: imageURL,
		}

		findIDArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			uid,
		}
		mockUserRepository.
			On("FindByID", findIDArgs...).
			Return(mockUser, nil)

		multipartImageFixture := fixture.NewMultipartImage("image.png", "image/png")
		defer multipartImageFixture.Close()
		imageFileHeader := multipartImageFixture.GetFormFile()
		imageFile, _ := imageFileHeader.Open()

		updateProfileArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mock.AnythingOfType("string"),
			imageFile,
		}

		mockImageRepository.
			On("UpdateProfile", updateProfileArgs...).
			Return(imageURL, nil)

		updateImageArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mockUser.UID,
			imageURL,
		}

		mockUpdatedUser := &domain.User{
			UID:      uid,
			Email:    "new@bob.com",
			Website:  "https://bob.com",
			Name:     "New Bob",
			ImageURL: imageURL,
		}

		mockUserRepository.
			On("UpdateImage", updateImageArgs...).
			Return(mockUpdatedUser, nil)

		ctx := context.Background()

		updatedUser, err := us.SetProfileImage(ctx, mockUser.UID, imageFileHeader)

		assert.NoError(t, err)
		assert.Equal(t, mockUpdatedUser, updatedUser)

		mockUserRepository.AssertCalled(t, "FindByID", findIDArgs...)
		mockImageRepository.AssertCalled(t, "UpdateProfile", updateProfileArgs...)
		mockUserRepository.AssertCalled(t, "UpdateImage", updateImageArgs...)
	})

	t.Run("UserRepository FindID Error", func(t *testing.T) {
		t.Parallel()

		uid, _ := uuid.NewRandom()

		findIDArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			uid,
		}
		mockErr := apperrors.NewInternal()
		mockUserRepository.
			On("FindByID", findIDArgs...).
			Return(nil, mockErr)

		multipartImageFixture := fixture.NewMultipartImage("image.png", "image/png")
		defer multipartImageFixture.Close()
		imageFileHeader := multipartImageFixture.GetFormFile()

		ctx := context.Background()

		updatedUser, err := us.SetProfileImage(ctx, uid, imageFileHeader)

		assert.Nil(t, updatedUser)
		assert.EqualError(t, err, mockErr.Error())

		mockUserRepository.AssertCalled(t, "FindByID", findIDArgs...)
		mockImageRepository.AssertNotCalled(t, "UpdateProfile")
		mockUserRepository.AssertNotCalled(t, "UpdateImage")
	})

	t.Run("ImageRepository Error", func(t *testing.T) {
		t.Parallel()

		// need to create a new UserService and repository
		// because testify has no way to overwrite a mock's
		// "On" call.
		_mockUserRepository := new(mocks.MockUserRepository)
		_mockImageRepository := new(mocks.MockImageRepository)

		_us := NewUserService(&USConfig{
			UserRepository:  _mockUserRepository,
			ImageRepository: _mockImageRepository,
		})

		uid, _ := uuid.NewRandom()
		imageURL := "http://imageurl.com/jdfkj34kljl"

		// has imageURL
		mockUser := &domain.User{
			UID:      uid,
			Email:    "bob@bob.com",
			Website:  "https://bob.com",
			Name:     "Bobby Bob",
			ImageURL: imageURL,
		}

		findIDArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			uid,
		}
		_mockUserRepository.
			On("FindByID", findIDArgs...).
			Return(mockUser, nil)

		multipartImageFixture := fixture.NewMultipartImage("image.png", "image/png")
		defer multipartImageFixture.Close()
		imageFileHeader := multipartImageFixture.GetFormFile()
		imageFile, _ := imageFileHeader.Open()

		updateProfileArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mock.AnythingOfType("string"),
			imageFile,
		}

		mockError := apperrors.NewInternal()
		_mockImageRepository.
			On("UpdateProfile", updateProfileArgs...).
			Return(nil, mockError)

		ctx := context.Background()

		updatedUser, err := _us.SetProfileImage(ctx, mockUser.UID, imageFileHeader)

		assert.Nil(t, updatedUser)
		assert.EqualError(t, err, mockError.Error())

		_mockUserRepository.AssertCalled(t, "FindByID", findIDArgs...)
		_mockImageRepository.AssertCalled(t, "UpdateProfile", updateProfileArgs...)
		_mockUserRepository.AssertNotCalled(t, "UpdateImage")
	})

	t.Run("UserRepository UpdateImage Error", func(t *testing.T) {
		t.Parallel()

		uid, _ := uuid.NewRandom()
		imageURL := "http://imageurl.com/jdfkj34kljl"

		// has imageURL
		mockUser := &domain.User{
			UID:      uid,
			Email:    "bob@bob.com",
			Website:  "https://bob.com",
			Name:     "Bobby Bob",
			ImageURL: imageURL,
		}

		findIDArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			uid,
		}
		mockUserRepository.
			On("FindByID", findIDArgs...).
			Return(mockUser, nil)

		multipartImageFixture := fixture.NewMultipartImage("image.png", "image/png")
		defer multipartImageFixture.Close()
		imageFileHeader := multipartImageFixture.GetFormFile()
		imageFile, _ := imageFileHeader.Open()

		updateProfileArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mock.AnythingOfType("string"),
			imageFile,
		}

		mockImageRepository.
			On("UpdateProfile", updateProfileArgs...).
			Return(imageURL, nil)

		updateImageArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mockUser.UID,
			imageURL,
		}

		mockError := apperrors.NewInternal()
		mockUserRepository.
			On("UpdateImage", updateImageArgs...).
			Return(nil, mockError)

		ctx := context.Background()

		updatedUser, err := us.SetProfileImage(ctx, mockUser.UID, imageFileHeader)

		assert.Nil(t, updatedUser)
		assert.EqualError(t, err, mockError.Error())

		mockUserRepository.AssertCalled(t, "FindByID", findIDArgs...)
		mockImageRepository.AssertCalled(t, "UpdateProfile", updateProfileArgs...)
		mockUserRepository.AssertCalled(t, "UpdateImage", updateImageArgs...)
	})
}
