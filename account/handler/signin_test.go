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

func TestSignIn(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	l := logger.Get()

	// setup mock services, gin engine/router, handler layer
	mockUserService := new(mocks.MockUserService)
	mockTokenService := new(mocks.MockTokenService)

	router := gin.New()
	router.Use(ginzap.Ginzap(l, time.RFC3339, false))
	router.Use(ginzap.RecoveryWithZap(l, true))

	NewHandler(&Config{
		R:            router,
		UserService:  mockUserService,
		TokenService: mockTokenService,
	})

	t.Run("Bad request data", func(t *testing.T) {
		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with invalid fields
		reqBody, err := json.Marshal(gin.H{
			"email":    "notanemail",
			"password": "short",
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)

		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockUserService.AssertNotCalled(t, "SignIn")
		mockTokenService.AssertNotCalled(t, "NewPairFromUser")
	})

	t.Run("Error Returned from UserService.Signin", func(t *testing.T) {
		email := "bob@bob.com"
		password := "pwdoesnotmatch123"

		mockUSArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			&domain.User{Email: email, Password: password},
		}

		// so we can check for a known status code
		mockError := apperrors.NewAuthorization("invalid email/password combo")

		mockUserService.On("SignIn", mockUSArgs...).Return(mockError)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with valid fields
		reqBody, err := json.Marshal(gin.H{
			"email":    email,
			"password": password,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)

		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		mockUserService.AssertCalled(t, "SignIn", mockUSArgs...)
		mockTokenService.AssertNotCalled(t, "NewPairFromUser")
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Successful Token Creation", func(t *testing.T) {
		email := "bob@bob.com"
		password := "pwworksgreat123"

		mockUSArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			&domain.User{Email: email, Password: password},
		}

		mockUserService.On("SignIn", mockUSArgs...).Return(nil)

		mockTSArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			&domain.User{Email: email, Password: password},
			"",
		}

		mockTokenPair := &domain.TokenPair{
			IDToken:      domain.IDToken{SS: "idToken"},
			RefreshToken: domain.RefreshToken{SS: "refreshToken"},
		}

		mockTokenService.On("NewPairFromUser", mockTSArgs...).Return(mockTokenPair, nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with valid fields
		reqBody, err := json.Marshal(gin.H{
			"email":    email,
			"password": password,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)

		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{
			"tokens": mockTokenPair,
		})
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())

		mockUserService.AssertCalled(t, "SignIn", mockUSArgs...)
		mockTokenService.AssertCalled(t, "NewPairFromUser", mockTSArgs...)
	})

	t.Run("Failed Token Creation", func(t *testing.T) {
		email := "cannotproducetoken@bob.com"
		password := "cannotproducetoken"

		mockUSArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			&domain.User{Email: email, Password: password},
		}

		mockUserService.On("SignIn", mockUSArgs...).Return(nil)

		mockTSArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			&domain.User{Email: email, Password: password},
			"",
		}

		mockError := apperrors.NewInternal()
		mockTokenService.On("NewPairFromUser", mockTSArgs...).Return(nil, mockError)
		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with valid fields
		reqBody, err := json.Marshal(gin.H{
			"email":    email,
			"password": password,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/signin", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)

		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{
			"error": mockError,
		})
		assert.NoError(t, err)

		assert.Equal(t, mockError.Status(), rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())

		mockUserService.AssertCalled(t, "SignIn", mockUSArgs...)
		mockTokenService.AssertCalled(t, "NewPairFromUser", mockTSArgs...)
	})
}
