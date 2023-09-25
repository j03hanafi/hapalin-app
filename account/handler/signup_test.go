package handler

import (
	"bytes"
	"encoding/json"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
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

func TestSignUp(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	l := logger.Get()

	t.Run("Email and Password Required", func(t *testing.T) {
		t.Parallel()

		mockUserService := new(mocks.MockUserService)
		mockUserService.
			On("SignUp", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*domain.User")).
			Return(nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		reqBody, err := json.Marshal(gin.H{
			"email": "",
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockUserService.AssertNotCalled(t, "SignUp")
	})

	t.Run("Invalid email", func(t *testing.T) {
		t.Parallel()

		mockUserService := new(mocks.MockUserService)
		mockUserService.
			On("SignUp", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*domain.User")).
			Return(nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		reqBody, err := json.Marshal(gin.H{
			"email":    "bobbo@b",
			"password": "atLeast8Chars",
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockUserService.AssertNotCalled(t, "SignUp")
	})

	t.Run("Password too short", func(t *testing.T) {
		t.Parallel()

		mockUserService := new(mocks.MockUserService)
		mockUserService.
			On("SignUp", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*domain.User")).
			Return(nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		reqBody, err := json.Marshal(gin.H{
			"email":    "bob@bob.com",
			"password": "failed",
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockUserService.AssertNotCalled(t, "SignUp")
	})

	t.Run("Password too long", func(t *testing.T) {
		t.Parallel()

		mockUserService := new(mocks.MockUserService)
		mockUserService.
			On("SignUp", mock.AnythingOfType("*gin.Context"), mock.AnythingOfType("*domain.User")).
			Return(nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		reqBody, err := json.Marshal(gin.H{
			"email":    "bob@bob.com",
			"password": "This Password is Too Long and will Fail when validated and will return a 400",
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockUserService.AssertNotCalled(t, "SignUp")
	})

	t.Run("Error calling UserService", func(t *testing.T) {
		t.Parallel()

		u := &domain.User{
			Email:    "bob@bob.com",
			Password: "atLeast8Chars",
		}

		mockUserService := new(mocks.MockUserService)
		mockUserService.
			On("SignUp", mock.AnythingOfType("*gin.Context"), u).
			Return(apperrors.NewConflict("User already exists", u.Email))

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		NewHandler(&Config{
			R:           router,
			UserService: mockUserService,
		})

		reqBody, err := json.Marshal(gin.H{
			"email":    u.Email,
			"password": u.Password,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusConflict, rr.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("Successful token creation", func(t *testing.T) {
		t.Parallel()

		u := &domain.User{
			Email:    "bob@bob.com",
			Password: "atLeast8Chars",
		}

		mockTokenResp := &domain.TokenPair{
			IDToken:      "idToken",
			RefreshToken: "refreshToken",
		}

		mockUserService := new(mocks.MockUserService)
		mockTokenService := new(mocks.MockTokenService)

		mockUserService.
			On("SignUp", mock.AnythingOfType("*gin.Context"), u).
			Return(nil)
		mockTokenService.
			On("NewPairFromUser", mock.AnythingOfType("*gin.Context"), u, "").
			Return(mockTokenResp, nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		NewHandler(&Config{
			R:            router,
			UserService:  mockUserService,
			TokenService: mockTokenService,
		})

		reqBody, err := json.Marshal(gin.H{
			"email":    u.Email,
			"password": u.Password,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{
			"tokens": mockTokenResp,
		})
		assert.NoError(t, err)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())

		mockUserService.AssertExpectations(t)
		mockTokenService.AssertExpectations(t)
	})

	t.Run("Failed token creation", func(t *testing.T) {
		t.Parallel()

		u := &domain.User{
			Email:    "bob@bob.com",
			Password: "atLeast8Chars",
		}

		mockErrorResp := apperrors.NewInternal()

		mockUserService := new(mocks.MockUserService)
		mockTokenService := new(mocks.MockTokenService)

		mockUserService.
			On("SignUp", mock.AnythingOfType("*gin.Context"), u).
			Return(nil)
		mockTokenService.
			On("NewPairFromUser", mock.AnythingOfType("*gin.Context"), u, "").
			Return(nil, mockErrorResp)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		router := gin.New()
		router.Use(ginzap.Ginzap(l, time.RFC3339, false))
		router.Use(ginzap.RecoveryWithZap(l, true))

		NewHandler(&Config{
			R:            router,
			UserService:  mockUserService,
			TokenService: mockTokenService,
		})

		reqBody, err := json.Marshal(gin.H{
			"email":    u.Email,
			"password": u.Password,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		request.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{
			"error": mockErrorResp,
		})
		assert.NoError(t, err)

		assert.Equal(t, mockErrorResp.Status(), rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())

		mockUserService.AssertExpectations(t)
		mockTokenService.AssertExpectations(t)
	})
}
