package handler

import (
	"encoding/json"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/domain/fixture"
	"github.com/j03hanafi/hapalin-app/account/domain/mocks"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestImage(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	l := logger.Get()

	uid, _ := uuid.NewRandom()
	ctxUser := domain.User{
		UID: uid,
	}

	router := gin.New()
	router.Use(ginzap.Ginzap(l, time.RFC3339, false))
	router.Use(ginzap.RecoveryWithZap(l, true))
	router.Use(func(c *gin.Context) {
		c.Set("user", &ctxUser)
	})

	mockUserService := new(mocks.MockUserService)

	NewHandler(&Config{
		R:            router,
		UserService:  mockUserService,
		MaxBodyBytes: 4 << 20, // 4 MiB
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()

		imageURL := "https://www.ImageURL.com/1234"

		multipartImageFixture := fixture.NewMultipartImage("image.png", "image/png")
		defer multipartImageFixture.Close()

		setProfileImageArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			ctxUser.UID,
			mock.AnythingOfType("*multipart.FileHeader"),
		}

		updatedUser := ctxUser
		updatedUser.ImageURL = imageURL

		mockUserService.
			On("SetProfileImage", setProfileImageArgs...).
			Return(&updatedUser, nil)

		request, _ := http.NewRequest(http.MethodPost, "/image", multipartImageFixture.MultipartBody)
		request.Header.Set("Content-Type", multipartImageFixture.ContentType)

		router.ServeHTTP(rr, request)

		respBody, _ := json.Marshal(gin.H{
			"imageUrl": imageURL,
			"message":  "Profile image updated successfully",
		})

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())

		mockUserService.AssertCalled(t, "SetProfileImage", setProfileImageArgs...)
	})

	t.Run("Disallowed mimetype", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()

		multipartImageFixture := fixture.NewMultipartImage("image.txt", "mage/svg+xml")
		defer multipartImageFixture.Close()

		request, _ := http.NewRequest(http.MethodPost, "/image", multipartImageFixture.MultipartBody)
		request.Header.Set("Content-Type", "multipart/form-data")

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		mockUserService.AssertNotCalled(t, "SetProfileImage")
	})

	t.Run("No image file provided", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()

		request, _ := http.NewRequest(http.MethodPost, "/image", nil)
		request.Header.Set("Content-Type", "multipart/form-data")

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		mockUserService.AssertNotCalled(t, "SetProfileImage")
	})

	t.Run("Error from SetProfileImage", func(t *testing.T) {
		t.Parallel()
		// create unique context user for this test
		uid, _ := uuid.NewRandom()
		ctxUser := domain.User{
			UID: uid,
		}

		router := gin.Default()
		router.Use(func(c *gin.Context) {
			c.Set("user", &ctxUser)
		})

		mockUserService := new(mocks.MockUserService)

		NewHandler(&Config{
			R:            router,
			UserService:  mockUserService,
			MaxBodyBytes: 4 * 1024 * 1024,
		})

		rr := httptest.NewRecorder()

		multipartImageFixture := fixture.NewMultipartImage("image.png", "image/png")
		defer multipartImageFixture.Close()

		setProfileImageArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			ctxUser.UID,
			mock.AnythingOfType("*multipart.FileHeader"),
		}

		mockError := apperrors.NewInternal()

		mockUserService.On("SetProfileImage", setProfileImageArgs...).Return(nil, mockError)

		request, _ := http.NewRequest(http.MethodPost, "/image", multipartImageFixture.MultipartBody)
		request.Header.Set("Content-Type", multipartImageFixture.ContentType)

		router.ServeHTTP(rr, request)

		assert.Equal(t, apperrors.Status(mockError), rr.Code)

		mockUserService.AssertCalled(t, "SetProfileImage", setProfileImageArgs...)
	})
}
