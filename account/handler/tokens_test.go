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
	gin.SetMode(gin.TestMode)
	l := logger.Get()

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
		t.Parallel()

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with invalid fields
		reqBody, _ := json.Marshal(gin.H{
			"notRefreshToken": "this key is not valid for this handler!",
		})

		request, _ := http.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockTokenService.AssertNotCalled(t, "ValidateRefreshToken")
		mockUserService.AssertNotCalled(t, "Get")
		mockTokenService.AssertNotCalled(t, "NewPairFromUser")
	})

	t.Run("Invalid token", func(t *testing.T) {
		t.Parallel()

		invalidTokenString := "invalid"
		mockErrorMessage := "authProbs"
		mockError := apperrors.NewAuthorization(mockErrorMessage)

		mockTokenService.
			On("ValidateRefreshToken", invalidTokenString).
			Return(nil, mockError)

		// a response recorder for getting written http response
		rr := httptest.NewRecorder()

		// create a request body with invalid fields
		reqBody, _ := json.Marshal(gin.H{
			"refresh_token": invalidTokenString,
		})

		request, _ := http.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		respBody, _ := json.Marshal(gin.H{
			"error": mockError,
		})

		assert.Equal(t, mockError.Status(), rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
		mockTokenService.AssertCalled(t, "ValidateRefreshToken", invalidTokenString)
		mockUserService.AssertNotCalled(t, "Get")
		mockTokenService.AssertNotCalled(t, "NewPairFromUser")
	})

	t.Run("Failure to create new token pair", func(t *testing.T) {
		t.Parallel()

		validTokenString := "valid"
		mockTokenID, _ := uuid.NewRandom()
		mockUserID, _ := uuid.NewRandom()

		mockRefreshTokenResp := &domain.RefreshToken{
			SS:  validTokenString,
			ID:  mockTokenID,
			UID: mockUserID,
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

		mockError := apperrors.NewAuthorization("Invalid refresh token")
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

		// create a request body with invalid fields
		reqBody, _ := json.Marshal(gin.H{
			"refresh_token": validTokenString,
		})

		request, _ := http.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		respBody, _ := json.Marshal(gin.H{
			"error": mockError,
		})

		assert.Equal(t, mockError.Status(), rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
		mockTokenService.AssertCalled(t, "ValidateRefreshToken", validTokenString)
		mockUserService.AssertCalled(t, "Get", getArgs...)
		mockTokenService.AssertCalled(t, "NewPairFromUser", newPairArgs...)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		validTokenString := "anothervalid"
		mockTokenID, _ := uuid.NewRandom()
		mockUserID, _ := uuid.NewRandom()

		mockRefreshTokenResp := &domain.RefreshToken{
			SS:  validTokenString,
			ID:  mockTokenID,
			UID: mockUserID,
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
			IDToken: domain.IDToken{SS: "aNewIDToken"},
			RefreshToken: domain.RefreshToken{
				SS:  "aNewRefreshToken",
				ID:  mockNewTokenID,
				UID: mockNewUserID,
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

		// create a request body with invalid fields
		reqBody, _ := json.Marshal(gin.H{
			"refresh_token": validTokenString,
		})

		request, _ := http.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(reqBody))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(rr, request)

		respBody, _ := json.Marshal(gin.H{
			"tokens": mockTokenPairResp,
		})

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, respBody, rr.Body.Bytes())
		mockTokenService.AssertCalled(t, "ValidateRefreshToken", validTokenString)
		mockUserService.AssertCalled(t, "Get", getArgs...)
		mockTokenService.AssertCalled(t, "NewPairFromUser", newPairArgs...)
	})

	// TODO - User not found
}
