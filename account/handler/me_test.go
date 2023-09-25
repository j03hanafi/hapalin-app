package handler

import (
	"encoding/json"
	"fmt"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/j03hanafi/hapalin-app/account/domain"
	"github.com/j03hanafi/hapalin-app/account/domain/apperrors"
	"github.com/j03hanafi/hapalin-app/account/domain/mocks"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMe(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	l := logger.Get()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		uid, _ := uuid.NewRandom()

		mockUserResp := &domain.User{
			UID:   uid,
			Email: "bob@bob.com",
			Name:  "Bobby Bobson",
		}

		mockUserService := new(mocks.MockUserService)
		mockUserService.
			On("Get", mock.AnythingOfType("context.backgroundCtx"), uid).
			Return(mockUserResp, nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		router.Use(func(c *gin.Context) {
			c.Set("user", &domain.User{
				UID: uid,
			})
		})

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		request, err := http.NewRequest(http.MethodGet, "/me", nil)
		assert.NoError(t, err)

		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{"user": mockUserResp})
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
		mockUserService.AssertExpectations(t)
	})

	t.Run("No user in context", func(t *testing.T) {
		t.Parallel()
		mockUserService := new(mocks.MockUserService)
		mockUserService.
			On("Get", mock.Anything, mock.Anything).
			Return(nil, nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		request, err := http.NewRequest(http.MethodGet, "/me", nil)
		assert.NoError(t, err)

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockUserService.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
	})

	t.Run("User not found", func(t *testing.T) {
		t.Parallel()
		uid, _ := uuid.NewRandom()
		mockUserService := new(mocks.MockUserService)
		mockUserService.
			On("Get", mock.Anything, uid).
			Return(nil, fmt.Errorf("some error down call chain"))

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		router.Use(func(c *gin.Context) {
			c.Set("user", &domain.User{
				UID: uid,
			})
		})

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		request, err := http.NewRequest(http.MethodGet, "/me", nil)
		assert.NoError(t, err)

		router.ServeHTTP(rr, request)

		respErr := apperrors.NewNotFound("user", uid.String())
		respBody, err := json.Marshal(gin.H{"error": respErr})
		assert.NoError(t, err)

		assert.Equal(t, respErr.Status(), rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
		mockUserService.AssertExpectations(t)
	})
}
