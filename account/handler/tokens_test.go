package handler

import (
	"bytes"
	"encoding/json"
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

func TestTokens(t *testing.T) {
	t.Parallel()

	// Setup
	gin.SetMode(gin.TestMode)
	l := logger.Get()

	// Setup mock services, gin engine/router, handler layer
	mockTokenService := new(mocks.MockTokenService)
	mockUserService := new(mocks.MockUserService)

	router := gin.New()
	router.Use(ginzap.Ginzap(l, time.RFC3339, false))
	router.Use(ginzap.RecoveryWithZap(l, true))

	NewHandler(&Config{
		R:            router,
		TokenService: mockTokenService,
		UserService:  mockUserService,
	})

	t.Run("Invalid request", func(t *testing.T) {
		//t.Parallel()

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with invalid fields
		reqBody, err := json.Marshal(gin.H{
			"no_refresh_token": "invalidRefreshToken",
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)

		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		mockTokenService.AssertNotCalled(t, "ValidateRefreshToken")
		mockUserService.AssertNotCalled(t, "Get")
		mockTokenService.AssertNotCalled(t, "NewPairFromUser")
	})

	t.Run("Invalid token", func(t *testing.T) {
		//t.Parallel()

		invalidTokenString := "invalidToken"
		mockErrorMessage := "authProblem"
		mockErr := apperrors.NewAuthorization(mockErrorMessage)

		mockTokenService.
			On("ValidateRefreshToken", invalidTokenString).
			Return(nil, mockErr)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with invalid token
		reqBody, err := json.Marshal(gin.H{
			"refresh_token": invalidTokenString,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)

		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{
			"error": mockErr,
		})
		assert.NoError(t, err)

		assert.Equal(t, mockErr.Status(), rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())

		mockTokenService.AssertCalled(t, "ValidateRefreshToken", invalidTokenString)
		mockUserService.AssertNotCalled(t, "Get")
		mockTokenService.AssertNotCalled(t, "NewPairFromUser")
	})

	t.Run("Failure create a new token pair", func(t *testing.T) {
		//t.Parallel()

		validTokenString := "validToken"
		mockTokenID, _ := uuid.NewRandom()
		mockUserID, _ := uuid.NewRandom()
		mockRefreshTokenResp := &domain.RefreshToken{
			ID:  mockTokenID,
			UID: mockUserID,
			SS:  validTokenString,
		}
		mockTokenService.
			On("ValidateRefreshToken", validTokenString).
			Return(mockRefreshTokenResp, nil)

		mockUserResp := &domain.User{
			UID: mockUserID,
		}
		getArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mockRefreshTokenResp.UID,
		}
		mockUserService.
			On("Get", getArgs...).
			Return(mockUserResp, nil)

		mockError := apperrors.NewAuthorization("invalid refresh token")
		newPairArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mockUserResp,
			mockRefreshTokenResp.ID.String(),
		}
		mockTokenService.
			On("NewPairFromUser", newPairArgs...).
			Return(nil, mockError)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with valid token
		reqBody, err := json.Marshal(gin.H{
			"refresh_token": validTokenString,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)

		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{
			"error": mockError,
		})
		assert.NoError(t, err)

		assert.Equal(t, mockError.Status(), rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())

		mockTokenService.AssertCalled(t, "ValidateRefreshToken", validTokenString)
		mockUserService.AssertCalled(t, "Get", getArgs...)
		mockTokenService.AssertCalled(t, "NewPairFromUser", newPairArgs...)
	})

	t.Run("Success", func(t *testing.T) {
		//t.Parallel()

		validTokenString := "validToken"
		mockTokenID, _ := uuid.NewRandom()
		mockUserID, _ := uuid.NewRandom()
		mockRefreshTokenResp := &domain.RefreshToken{
			ID:  mockTokenID,
			UID: mockUserID,
			SS:  validTokenString,
		}
		mockTokenService.
			On("ValidateRefreshToken", validTokenString).
			Return(mockRefreshTokenResp, nil)

		mockUserResp := &domain.User{
			UID: mockUserID,
		}
		getArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mockRefreshTokenResp.UID,
		}
		mockUserService.
			On("Get", getArgs...).
			Return(mockUserResp, nil)

		mockNewTokenID, _ := uuid.NewRandom()
		mockNewUserID, _ := uuid.NewRandom()
		mockTokenPairResp := &domain.TokenPair{
			IDToken: domain.IDToken{
				SS: "aNewIDToken",
			},
			RefreshToken: domain.RefreshToken{
				ID:  mockNewTokenID,
				UID: mockNewUserID,
				SS:  "aNewRefreshToken",
			},
		}
		newPairArgs := mock.Arguments{
			mock.AnythingOfType("context.backgroundCtx"),
			mockUserResp,
			mockRefreshTokenResp.ID.String(),
		}
		mockTokenService.
			On("NewPairFromUser", newPairArgs...).
			Return(mockTokenPairResp, nil)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with valid token
		reqBody, err := json.Marshal(gin.H{
			"refresh_token": validTokenString,
		})
		assert.NoError(t, err)

		request, err := http.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)

		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		respBody, err := json.Marshal(gin.H{
			"tokens": mockTokenPairResp,
		})
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())

		mockTokenService.AssertCalled(t, "ValidateRefreshToken", validTokenString)
		mockUserService.AssertCalled(t, "Get", getArgs...)
		mockTokenService.AssertCalled(t, "NewPairFromUser", newPairArgs...)
	})

}
